package main

import (
	"bytes"
	"io"
)

type Parser interface {
	Prep(bs []byte) []byte
	Parse(bs []byte) (before, after []byte, found bool)
}

type IndexByteParser struct{}

func (p *IndexByteParser) Prep(bs []byte) []byte {
	return bs
}
func (p *IndexByteParser) Parse(bs []byte) (before, after []byte, found bool) {
	// '\r', '\n' 을 각각 전체 검색하면 한쪽이 없는 스트림에서 라인마다 남은 버퍼를
	// 끝까지 헛스캔한다(청크 안에서 O(n^2)). 첫 구분자가 '\r' 이라면 반드시 '\n' 보다
	// 앞에 있으므로, '\n' 을 먼저 찾고 '\r' 탐색을 그 앞 구간으로 제한해도 결과는 같다.
	//
	// 단 '\r' 만 나오는 스트림(curl progress)은 '\n' 탐색이 여전히 전체를 훑는다.
	// 양쪽을 모두 없애려면 창 단위로 두 구분자를 함께 찾아야 한다.
	indexN := bytes.IndexByte(bs, bn)
	if indexN == -1 {
		indexR := bytes.IndexByte(bs, br)
		if indexR == -1 {
			return before, after, false
		}
		return bs[:indexR], bs[indexR+1:], true
	}

	// 의도된 '\n\n' 은 그대로 출력하고, '\r\n' or '\n\r'은 '\n' 으로 치환해서 불필요한 줄바꿈 보정
	if indexR := bytes.IndexByte(bs[:indexN], br); indexR != -1 {
		cnt := 1
		if indexR+1 == indexN { // "\r\n"
			cnt = 2
		}
		return bs[:indexR], bs[indexR+cnt:], true
	}

	cnt := 1
	if indexN+1 < len(bs) && bs[indexN+1] == br { // "\n\r"
		cnt = 2
	}
	return bs[:indexN], bs[indexN+cnt:], true
}

// IndexByteParser는 '\r' 로만 끝나는 스트림(curl progress)에서 라인마다 '\n' 을 버퍼 끝까지 헛스캔한다.
// WindowParser는 일정 크기 창 안에서 두 구분자를 함께 찾아 헛스캔을 창 하나로 묶는다.
// 대신 라인이 아주 길면 창 단위 반복이 SIMD 한 번보다 불리하다.
type WindowParser struct{}

func (p *WindowParser) Prep(bs []byte) []byte {
	return bs
}
func (p *WindowParser) Parse(bs []byte) (before, after []byte, found bool) {
	for start := 0; start < len(bs); start += scanWindowLength {
		seg := bs[start:min(start+scanWindowLength, len(bs))]
		indexR := bytes.IndexByte(seg, br)
		indexN := bytes.IndexByte(seg, bn)
		if indexR == -1 && indexN == -1 {
			continue
		}
		i := indexR
		if indexR == -1 || (indexN != -1 && indexN < indexR) {
			i = indexN
		}
		index := start + i

		// 의도된 '\n\n' 은 그대로 출력하고, '\r\n' or '\n\r'은 '\n' 으로 치환해서 불필요한 줄바꿈 보정
		cnt := 1
		if index+1 < len(bs) {
			next := bs[index+1]
			if (next == br || next == bn) && next != bs[index] {
				cnt = 2
			}
		}
		return bs[:index], bs[index+cnt:], true
	}
	return before, after, false
}

type CutsParser struct{}

func (p *CutsParser) Prep(bs []byte) []byte {
	return bs
}
func (p *CutsParser) Parse(bs []byte) (before, after []byte, found bool) {
	beforeR, afterR, foundR := bytes.Cut(bs, bsr)
	beforeN, afterN, foundN := bytes.Cut(bs, bsn)
	if !foundR && !foundN {
		return before, after, false
	}
	before, after = beforeR, afterR
	if !foundR || (foundN && len(beforeN) < len(beforeR)) {
		before, after = beforeN, afterN
	}
	return before, after, true
}

type SliceParser struct{}

func (p *SliceParser) Prep(bs []byte) []byte {
	return bs
}
func (p *SliceParser) Parse(bs []byte) (before, after []byte, found bool) {
	for i := range bs {
		if bs[i] == '\r' || bs[i] == '\n' {
			return bs[:i], bs[i+1:], true
		}
	}
	return before, after, false
}

type IndexAnyParser struct{}

func (p *IndexAnyParser) Prep(bs []byte) []byte {
	return bs
}
func (p *IndexAnyParser) Parse(bs []byte) (before, after []byte, found bool) {
	index := bytes.IndexAny(bs, "\r\n")
	if index == -1 {
		return before, after, false
	}
	return bs[:index], bs[index+1:], true
}

type ReplaceCutParser struct{}

func (p *ReplaceCutParser) Prep(bs []byte) []byte {
	return bytes.ReplaceAll(bs, bsr, bsn)
	// 의도된 '\n\n' 도 치환되버릴수 있음
	// bs = bytes.ReplaceAll(chunk, bnn, bn)
}
func (p *ReplaceCutParser) Parse(bs []byte) (before, after []byte, found bool) {
	return bytes.Cut(bs, bsn)
}

type ReplaceSplitParser struct {
	split [][]byte
	index int
}

func (p *ReplaceSplitParser) Prep(bs []byte) []byte {
	return bytes.ReplaceAll(bs, bsr, bsn)
}
func (p *ReplaceSplitParser) Parse(bs []byte) (before, after []byte, found bool) {
	if len(p.split) == 0 {
		p.split = bytes.Split(bs, bsn)
	}
	before, after = p.split[p.index], p.split[len(p.split)-1]
	p.index++
	if p.index >= len(p.split) {
		p.split = [][]byte{}
		p.index = 0
		return before, after, false
	}
	return before, after, true
}

func parse(dst io.Writer, src io.Reader, p Parser, prefix string) {
	buf := make([]byte, readBufLength)
	stream := bytes.Buffer{}
	out := bytes.Buffer{}
	bprefix := []byte(prefix)

	for {
		n, err := src.Read(buf)
		if n > 0 {
			chunk := p.Prep(buf[:n])
			stream.Write(chunk)
			sBytes := stream.Bytes()

			// 예를 들어 "12\n34\n5" 중 "12", "34"는 각각의 라인으로 잘라서 전송하고
			for {
				before, after, found := p.Parse(sBytes)
				if !found {
					break
				}
				// // 추가시 의도된 '\n\n'도 치환되버림
				// if len(before) > 0 {
				// 	writeLine(&out, bprefix, before)
				// }
				writeLine(&out, bprefix, before)
				sBytes = after
			}

			// 마지막 "5"는 아직 라인이 미완성이므로 버퍼에 남겨둠
			if stream.Len() != len(sBytes) {
				stream.Reset()
				stream.Write(sBytes)
			}

			// chunk가 '\r' or '\n' 없이 계속 들어올때 stream 무한 증가하지 않게 강제로 라인 Write
			if stream.Len() > maxLineLength {
				writeLine(&out, bprefix, stream.Bytes())
				stream.Reset()
			}

			flushLines(dst, &out)
		}

		if err != nil {
			// '\n' 없이 끝난 경우 강제로 라인 Write
			if stream.Len() > 0 {
				writeLine(&out, bprefix, stream.Bytes())
			}
			flushLines(dst, &out)
			break
		}
	}
}

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

func (p IndexByteParser) Prep(bs []byte) []byte {
	return bs
}
func (p IndexByteParser) Parse(bs []byte) (before, after []byte, found bool) {
	// '\r', '\n' 둘 다 검색
	indexR := bytes.IndexByte(bs, br)
	indexN := bytes.IndexByte(bs, bn)
	if indexR == -1 && indexN == -1 {
		return before, after, false
	}
	index := indexR
	if indexR == -1 || (indexN > -1 && indexN < indexR) {
		index = indexN
	}
	// 의도된 '\n\n' 은 그대로 출력하고, '\r\n' or '\n\r'은 '\n' 으로 치환해서 불필요한 줄바꿈 보정
	cnt := 1
	if indexR != -1 && indexN != -1 && (indexR-indexN == 1 || indexR-indexN == -1) {
		cnt = 2
	}
	return bs[:index], bs[index+cnt:], true
}

type CutsParser struct{}

func (p CutsParser) Prep(bs []byte) []byte {
	return bs
}
func (p CutsParser) Parse(bs []byte) (before, after []byte, found bool) {
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

func (p SliceParser) Prep(bs []byte) []byte {
	return bs
}
func (p SliceParser) Parse(bs []byte) (before, after []byte, found bool) {
	for i := 0; i < len(bs); i++ {
		if bs[i] == '\r' || bs[i] == '\n' {
			return bs[:i], bs[i+1:], true
		}
	}
	return before, after, false
}

type IndexAnyParser struct{}

func (p IndexAnyParser) Prep(bs []byte) []byte {
	return bs
}
func (p IndexAnyParser) Parse(bs []byte) (before, after []byte, found bool) {
	index := bytes.IndexAny(bs, "\r\n")
	if index == -1 {
		return before, after, false
	}
	return bs[:index], bs[index+1:], true
}

type ReplaceCutParser struct{}

func (p ReplaceCutParser) Prep(bs []byte) []byte {
	return bytes.ReplaceAll(bs, bsr, bsn)
	// 의도된 '\n\n' 도 치환되버릴수 있음
	// bs = bytes.ReplaceAll(chunk, bnn, bn)
}
func (p ReplaceCutParser) Parse(bs []byte) (before, after []byte, found bool) {
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
	buf := make([]byte, 4096)
	stream := &bytes.Buffer{}
	line := &bytes.Buffer{}
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
				// 	dst.Write(concatBytes(line, bprefix, before, bsn))
				// }
				dst.Write(concatBytes(line, bprefix, before, bsn))
				sBytes = after
			}

			// 마지막 "5"는 아직 라인이 미완성이므로 버퍼에 남겨둠
			if stream.Len() != len(sBytes) {
				stream.Reset()
				stream.Write(sBytes)
			}

			// chunk가 '\r' or '\n' 없이 계속 들어올때 out 무한 증가하지 않게 강제로 라인 Write
			if stream.Len() > maxLineLength {
				dst.Write(concatBytes(line, bprefix, stream.Bytes(), bsn))
				stream.Reset()
			}
		}

		if err != nil {
			// '\n' 없이 끝난 경우 강제로 라인 Write
			if stream.Len() > 0 {
				dst.Write(concatBytes(line, bprefix, stream.Bytes(), bsn))
			}
			break
		}
	}
}

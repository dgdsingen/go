// parser.go = 현재 사용되는 parser
// bytes.Cut()만 n회 호출하는 버전
// 성능과 메모리 효율성이 가장 뛰어남
package main

import (
	"bytes"
	"io"
)

func parseCuts(dst io.Writer, src io.Reader, prefix string) {
	buf := make([]byte, 4096)
	stream := new(bytes.Buffer)
	line := new(bytes.Buffer)
	bprefix := []byte(prefix)

	for {
		n, err := src.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			// 예를 들어 "12\n34\n5" 중 "12", "34"는 각각의 라인으로 잘라서 전송하고
			for {
				// '\r', '\n' 둘 다 검색
				beforeR, afterR, foundR := bytes.Cut(chunk, br)
				beforeN, afterN, foundN := bytes.Cut(chunk, bn)
				if !foundR && !foundN {
					break
				}
				// '\r', '\n' 중 검색된 쪽을 선택. 둘 다 검색되었다면 더 앞쪽에 있는 것을 선택.
				before := beforeR
				chunk = afterR
				if !foundR || (foundN && len(beforeN) < len(beforeR)) {
					before, chunk = beforeN, afterN
				}
				// '\r', '\n' 가 문자열 가장 앞에 있었다면 skip
				if len(before) <= 0 {
					continue
				}
				// stream에 미완성 라인 잔여물이 남아있다면 함께 전송
				if stream.Len() > 0 {
					stream.Write(before)
					before = stream.Bytes()
					stream.Reset()
				}
				dst.Write(concatBytes(line, bprefix, before, bn))
			}

			// 마지막 "5"는 아직 라인이 미완성이므로 버퍼에 남겨둠
			stream.Write(chunk)

			// chunk가 '\r' or '\n' 없이 계속 들어올때 out 무한 증가하지 않게 강제로 라인 Write
			if stream.Len() > maxLineLength {
				dst.Write(concatBytes(line, bprefix, stream.Bytes(), bn))
				stream.Reset()
			}
		}

		if err != nil {
			// '\n' 없이 끝난 경우 강제로 라인 Write
			if stream.Len() > 0 {
				dst.Write(concatBytes(line, bprefix, stream.Bytes(), bn))
			}
			break
		}
	}
}

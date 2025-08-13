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
			stream.Write(buf[:n])
			sBytes := stream.Bytes()

			// 예를 들어 "12\n34\n5" 중 "12", "34"는 각각의 라인으로 잘라서 전송하고
			for {
				// '\r', '\n' 둘 다 검색
				beforeR, afterR, foundR := bytes.Cut(sBytes, br)
				beforeN, afterN, foundN := bytes.Cut(sBytes, bn)
				if !foundR && !foundN {
					break
				}
				before, after := beforeR, afterR
				if !foundR || (foundN && len(beforeN) < len(beforeR)) {
					before, after = beforeN, afterN
				}
				// TODO: 의도된 '\n\n' 도 치환되버릴수 있고, 반대로 불필요한 '\n' 가 출력될수도 있고.
				// if len(before) > 0 {
				dst.Write(concatBytes(line, bprefix, before, bn))
				// }
				sBytes = after
			}

			// 마지막 "5"는 아직 라인이 미완성이므로 버퍼에 남겨둠
			if stream.Len() != len(sBytes) {
				stream.Reset()
				stream.Write(sBytes)
			}

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

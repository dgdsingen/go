// parser.go = 현재 사용되는 parser
// bytes.Cut()만 n회 호출하는 버전. bytes.IndexByte() 버전 다음으로 뛰어남.
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
				beforeR, afterR, foundR := bytes.Cut(sBytes, bsr)
				beforeN, afterN, foundN := bytes.Cut(sBytes, bsn)
				if !foundR && !foundN {
					break
				}
				before, after := beforeR, afterR
				if !foundR || (foundN && len(beforeN) < len(beforeR)) {
					before, after = beforeN, afterN
				}
				// TODO: if len(before) > 0 추가시 의도된 '\n\n'도 치환되버림. if를 빼면 불필요한 '\n'가 출력될수도 있음.
				// '\r', '\n' 둘 다 찾았을때 before 길이 차이가 1인 경우 1개는 skip 처리한다면?
				if len(before) > 0 {
					dst.Write(concatBytes(line, bprefix, before, bsn))
				}
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

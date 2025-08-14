// Current Parser
// bytes.IndexByte() 버전
// bytes.Cut() 버전보다 효율적임
package main

import (
	"bytes"
	"io"
)

func parseIndexByte(dst io.Writer, src io.Reader, prefix string) {
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
				foundR := bytes.IndexByte(sBytes, br)
				foundN := bytes.IndexByte(sBytes, bn)
				if foundR == -1 && foundN == -1 {
					break
				}
				found := foundR
				if foundR == -1 || (foundN > -1 && foundN < foundR) {
					found = foundN
				}
				before, after := sBytes[:found], sBytes[found+1:]
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

// Deprecated
// slice 버전
package main

import (
	"bytes"
	"io"
)

func parseSlice(dst io.Writer, src io.Reader, prefix string) {
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
			for i := 0; i < len(sBytes); i++ {
				// '\r', '\n' 둘 다 검색
				if sBytes[i] == '\r' || sBytes[i] == '\n' {
					before, after := sBytes[:i], sBytes[i+1:]
					dst.Write(concatBytes(line, bprefix, before, bsn))
					sBytes, i = after, 0
				}
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

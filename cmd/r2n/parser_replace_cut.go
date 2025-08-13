// bytes.ReplaceAll() + bytes.Cut() 버전
// bytes.Cut()은 좋지만 bytes.ReplaceAll()가 성능과 메모리를 깎아먹음
package main

import (
	"bytes"
	"io"
)

func parseReplaceCut(dst io.Writer, src io.Reader, prefix string) {
	buf := make([]byte, 4096)
	stream := new(bytes.Buffer)
	line := new(bytes.Buffer)
	bprefix := []byte(prefix)

	for {
		n, err := src.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			chunk = bytes.ReplaceAll(chunk, br, bn)
			// 의도된 '\n\n' 도 치환되버릴수 있음
			// chunk = bytes.ReplaceAll(chunk, bnn, bn)
			stream.Write(chunk)

			// 예를 들어 "12\n34\n5" 중 "12", "34"는 각각의 라인으로 잘라서 전송하고
			sBytes := stream.Bytes()
			for {
				before, after, found := bytes.Cut(sBytes, bn)
				if !found {
					break
				}
				// 의도된 '\n\n' 도 치환되버릴수 있음
				// if len(before) > 0 {
				dst.Write(concatBytes(line, bprefix, before, bn))
				// }
				sBytes = after
			}

			// 마지막 "5"는 아직 라인이 미완성이므로 버퍼에 남겨둠
			stream.Reset()
			stream.Write(sBytes)

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

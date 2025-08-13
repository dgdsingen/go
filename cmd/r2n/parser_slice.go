// slice 버전
// slice 재할당이 너무 많이 일어나서 성능과 메모리 효율성 낮음
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
			chunk := buf[:n]
			// 예를 들어 "12\n34\n5" 중 "12", "34"는 각각의 라인으로 잘라서 전송하고
			for i := 0; i < len(chunk); i++ {
				// '\r', '\n' 둘 다 검색
				if chunk[i] == '\r' || chunk[i] == '\n' {
					before := chunk[:i]
					chunk = chunk[i+1:]
					i = 0
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

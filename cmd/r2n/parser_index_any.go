package main

import (
	"bytes"
	"io"
)

func copyAndReplaceIndexAny(dst io.Writer, src io.Reader, prefix string) {
	const maxLineLength = 64 * 1024 // 64KB

	buf := make([]byte, 4096)
	stream := new(bytes.Buffer)
	line := new(bytes.Buffer)

	bprefix := []byte(prefix)
	bn := []byte{'\n'}

	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.Write(buf[:n])

			// 예를 들어 "12\n34\n5" 중 "12", "34"는 각각의 라인으로 잘라서 전송하고
			sBytes := stream.Bytes()
			for {
				if found := bytes.IndexAny(sBytes, "\r\n"); found != -1 {
					before := sBytes[:found]
					after := sBytes[found+1:]
					if len(before) > 0 {
						dst.Write(concatBytes(line, bprefix, before, bn))
					}
					sBytes = after
				} else {
					break
				}
			}

			// 마지막 5는 아직 라인이 미완성이므로 버퍼에 남겨둠
			stream.Reset()
			stream.Write(sBytes)

			// chunk가 '\r' or '\n' 없이 계속 들어올때 out 무한 증가하지 않게 강제로 라인 Write
			if stream.Len() > maxLineLength {
				dst.Write(concatBytes(line, bprefix, stream.Bytes(), bn))
				stream.Reset()
			}
		}

		if err != nil {
			// '\n' 없이 끝난 경우 강제로 라인 처리해서 내보냄
			if stream.Len() > 0 {
				dst.Write(concatBytes(line, bprefix, stream.Bytes(), bn))
			}
			break
		}
	}
}

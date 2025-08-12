package main

import (
	"bytes"
	"io"
)

func copyAndReplaceSplit(dst io.Writer, src io.Reader, prefix string) {
	const maxLineLength = 64 * 1024 // 64KB

	buf := make([]byte, 4096)
	stream := new(bytes.Buffer)
	line := new(bytes.Buffer)

	bprefix := []byte(prefix)
	br := []byte{'\r'}
	bn := []byte{'\n'}
	bnn := []byte{'\n', '\n'}

	for {
		n, err := src.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			chunk = bytes.ReplaceAll(chunk, br, bn)
			chunk = bytes.ReplaceAll(chunk, bnn, bn)
			stream.Write(chunk)

			// chunk가 '\n' 없이 계속 들어올때 out 무한 증가를 막기 위해 강제 라인처리 + flush
			if stream.Len() > maxLineLength {
				dst.Write(concatBytes(line, bprefix, stream.Bytes(), bn))
				stream.Reset()
				continue
			}

			// 예를 들어 "12\n34\n5" 중 "12", "34"는 각각의 라인으로 잘라서 전송하고
			split := bytes.Split(stream.Bytes(), bn)
			for _, s := range split[:len(split)-1] {
				dst.Write(concatBytes(line, bprefix, s, bn))
			}

			// 마지막 5는 아직 라인이 미완성이므로 버퍼에 남겨둠
			last := split[len(split)-1]
			stream.Reset()
			stream.Write(last)
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

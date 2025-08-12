package main

import (
	"bytes"
	"io"
)

func copyAndReplaceSlice(dst io.Writer, src io.Reader, prefix string) {
	buf := make([]byte, 4096)
	// len > 0 이면 slice가 zero value로 채워져서 이상하게 출력될 수 있으므로 0으로 설정
	stream := make([]byte, 0, 4096)
	bprefix := []byte(prefix)

	for {
		n, err := src.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			chunk = replaceRN(chunk)
			stream = bytes.Join([][]byte{stream, chunk}, nil)

			// 예를 들어 "12\n34\n5" 중 "12", "34"는 각각의 라인으로 잘라서 전송하고
			p := 0
			for i, b := range stream {
				if b == '\n' {
					line := bytes.Join([][]byte{bprefix, stream[p : i+1]}, nil)
					dst.Write(line)
					p = i + 1
				}
			}

			// 마지막 5는 아직 라인이 미완성이므로 버퍼에 남겨둠
			if p < len(stream) {
				stream = stream[p:]
			} else {
				stream = stream[:0]
			}

			// chunk가 '\n' 없이 계속 들어올때 out 무한 증가를 막기 위해 강제 라인처리 + flush
			if len(stream) > maxLineLength {
				line := bytes.Join([][]byte{bprefix, stream, bn}, nil)
				dst.Write(line)
				stream = stream[:0]
			}
		}

		if err != nil {
			// '\n' 없이 끝난 경우 강제로 라인 처리해서 내보냄
			if len(stream) > 0 {
				line := bytes.Join([][]byte{bprefix, stream, bn}, nil)
				dst.Write(line)
			}
			break
		}
	}
}

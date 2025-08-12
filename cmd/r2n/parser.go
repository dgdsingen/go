package main

import (
	"bytes"
	"io"
)

func concatBytes(line *bytes.Buffer, bs ...[]byte) []byte {
	// system call을 줄이기 위해 라인 단위로 버퍼링해서 출력. 이게 bufio.Writer 보다 빠름
	defer line.Reset()
	for _, b := range bs {
		line.Write(b)
	}
	return line.Bytes()
}

func copyAndReplace(dst io.Writer, src io.Reader, prefix string) {
	const maxLineLength = 64 * 1024 // 64KB

	buf := make([]byte, 4096)
	stream := new(bytes.Buffer)
	line := new(bytes.Buffer)

	bprefix := []byte(prefix)
	bn := []byte{'\n'}
	br := []byte{'\r'}

	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.Write(buf[:n])

			// chunk가 '\r' or '\n' 없이 계속 들어올때 out 무한 증가하지 않게 강제로 라인 Write
			if stream.Len() > maxLineLength {
				dst.Write(concatBytes(line, bprefix, stream.Bytes(), bn))
				stream.Reset()
				continue
			}

			// 예를 들어 "12\n34\n5" 중 "12", "34"는 각각의 라인으로 잘라서 전송하고
			sBytes := stream.Bytes()
			for {
				// '\r' 로 먼저 Cut 해보고
				before, after, found := bytes.Cut(sBytes, br)
				if !found {
					// 못찾았으면 '\n' 로 다시 Cut
					before, after, found = bytes.Cut(sBytes, bn)
				}

				if found {
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

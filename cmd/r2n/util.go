package main

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
)

func fmtVersion() string {
	return fmt.Sprintf("%s %s", appName, version)
}

// func replaceRN(bs []byte) []byte {
// 	p := 0
// 	prev := byte(0)
// 	for _, b := range bs {
// 		if b == '\r' {
// 			b = '\n'
// 		}
// 		if b == '\n' && prev == '\n' {
// 			continue
// 		}
// 		bs[p] = b
// 		p++
// 		prev = b
// 	}
// 	return bs[:p]
// }

func writeLine(out *bytes.Buffer, prefix, line []byte) {
	out.Write(prefix)
	out.Write(line)
	out.WriteByte(bn)
}

func flushLines(dst io.Writer, out *bytes.Buffer) {
	if out.Len() == 0 {
		return
	}
	n, err := dst.Write(out.Bytes())
	if err != nil {
		slog.Error(err.Error(), slog.Int("n", n))
	}
	out.Reset()
}

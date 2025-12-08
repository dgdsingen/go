package main

import (
	"bytes"
	"fmt"
)

// fmt.Printf("----\n")
// fmt.Printf("%v %q(%d) %q(%d)\n", found, before, len(before), after, len(after))
// fmt.Printf("----\n")

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

func concatBytes(line *bytes.Buffer, bs ...[]byte) []byte {
	// system call을 줄이기 위해 라인 단위로 버퍼링해서 출력. 이게 bufio.Writer 보다 빠름
	defer line.Reset()
	for _, b := range bs {
		line.Write(b)
	}
	return line.Bytes()
}

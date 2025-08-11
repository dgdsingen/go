package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func BenchmarkCopyAndReplace(b *testing.B) {
	data := bytes.Repeat([]byte("Hello\r\nWorld\n\n"), 200000) // 10MB
	prefix := "[prefix] "
	for b.Loop() {
		src := bytes.NewReader(data)
		dst := io.Discard
		copyAndReplace(dst, src, &prefix)
	}
}

/*
r2n -prefix="[sh] " -stdio=stdout -- sh -c 'yes 1 | tr -d "\n" | head -c 100000'
과 같이 테스트 해봤는데 항상 4096B 에서 잘리는 것을 확인.
stdin이 파이프라서 src.Read(buf)가 보통 4096B 단위로 처리되는 듯.

아래와 같이 Go로 처리하면 제약없이 테스트 가능.
if out.Len() > maxLineLength 조건문이 없으면 out.Len()이 무한 증식하는 것이 확인된다.
*/
func TestLongLine(t *testing.T) {
	data := bytes.Repeat([]byte{'X'}, 100000) // 100KB without '\r' or '\n'
	src := bytes.NewReader([]byte(data))
	prefix := "[test] "
	copyAndReplace(os.Stdout, src, &prefix)
}

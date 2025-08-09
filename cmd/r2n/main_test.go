package main

import (
	"bytes"
	"testing"
)

func BenchmarkCopyAndReplace(b *testing.B) {
	data := bytes.Repeat([]byte("Hello\r\nWorld\n\n"), 200000) // 약 10MB
	prefix := "[prefix] "
	for b.Loop() {
		src := bytes.NewReader(data)
		var dst bytes.Buffer
		copyAndReplace(&dst, src, &prefix)
	}
}

package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

var (
	count  = 100000
	prefix = "[prefix] "
	// dst = new(bytes.Buffer)
	// dst = io.Discard
	dst = os.Stdout
)

// Benchmark warm-up
func BenchmarkWarmUp(b *testing.B) {
	data := bytes.Repeat(append(bytes.Repeat([]byte{'X'}, 10000), byte('\n')), count)
	src := bytes.NewReader(data)
	io.Copy(io.Discard, src)
}

func BenchmarkShortLines(b *testing.B) {
	data := bytes.Repeat(append(bytes.Repeat([]byte{'X'}, 100), byte('\n')), count)
	src := bytes.NewReader(data)
	copyAndReplace(dst, src, prefix)
}

func BenchmarkLongLines(b *testing.B) {
	data := bytes.Repeat(append(bytes.Repeat([]byte{'X'}, 10000), byte('\n')), count)
	src := bytes.NewReader(data)
	copyAndReplace(dst, src, prefix)
}

func BenchmarkMixedLines(b *testing.B) {
	data := bytes.Repeat([]byte("Hello\r\nWorld\n\n"), count)
	src := bytes.NewReader(data)
	copyAndReplace(dst, src, prefix)
}

func BenchmarkShortLinesSlice(b *testing.B) {
	data := bytes.Repeat(append(bytes.Repeat([]byte{'X'}, 100), byte('\n')), count)
	src := bytes.NewReader(data)
	copyAndReplaceSlice(dst, src, prefix)
}

func BenchmarkLongLinesSlice(b *testing.B) {
	data := bytes.Repeat(append(bytes.Repeat([]byte{'X'}, 10000), byte('\n')), count)
	src := bytes.NewReader(data)
	copyAndReplaceSlice(dst, src, prefix)
}

func BenchmarkMixedLinesSlice(b *testing.B) {
	data := bytes.Repeat([]byte("Hello\r\nWorld\n\n"), count)
	src := bytes.NewReader(data)
	copyAndReplaceSlice(dst, src, prefix)
}

func BenchmarkShortLinesCut(b *testing.B) {
	data := bytes.Repeat(append(bytes.Repeat([]byte{'X'}, 100), byte('\n')), count)
	src := bytes.NewReader(data)
	copyAndReplaceCut(dst, src, prefix)
}

func BenchmarkLongLinesCut(b *testing.B) {
	data := bytes.Repeat(append(bytes.Repeat([]byte{'X'}, 10000), byte('\n')), count)
	src := bytes.NewReader(data)
	copyAndReplaceCut(dst, src, prefix)
}

func BenchmarkMixedLinesCut(b *testing.B) {
	data := bytes.Repeat([]byte("Hello\r\nWorld\n\n"), count)
	src := bytes.NewReader(data)
	copyAndReplaceCut(dst, src, prefix)
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
	prefix := "[prefix] "
	copyAndReplace(os.Stdout, src, prefix)
}

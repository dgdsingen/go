package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// Benchmark 결과:
//   최대한 표준 라이브러리를 쓰자. bytes.Cuts() 성능이 평균적으로 가장 좋다.
//   slice 재할당은 성능과 메모리에 치명적이다.
//
// BenchmarkLongLinesParseCuts-16           1000000000    0.2180 ns/op   0 B/op   0 allocs/op
// BenchmarkLongLinesParseReplaceCut-16     1000000000    0.3692 ns/op   2 B/op   0 allocs/op
// BenchmarkLongLinesParseReplaceSplit-16   1000000000    0.4244 ns/op   2 B/op   0 allocs/op
// BenchmarkLongLinesParseIndexAny-16       1000000000    0.5914 ns/op   0 B/op   0 allocs/op
// BenchmarkLongLinesParseSlice-16          1000000000    0.8247 ns/op   0 B/op   0 allocs/op
// BenchmarkLongLinesParseScanner-16        1000000000    0.9255 ns/op   0 B/op   0 allocs/op
//
// BenchmarkMixedLinesParseSlice-16         1000000000    0.1205 ns/op   0 B/op   0 allocs/op
// BenchmarkMixedLinesParseIndexAny-16      1000000000    0.1251 ns/op   0 B/op   0 allocs/op
// BenchmarkMixedLinesParseCuts-16          1000000000    0.1284 ns/op   0 B/op   0 allocs/op
// BenchmarkMixedLinesParseReplaceCut-16    1000000000    0.1303 ns/op   0 B/op   0 allocs/op
// BenchmarkMixedLinesParseReplaceSplit-16  1000000000    0.1310 ns/op   0 B/op   0 allocs/op
// BenchmarkMixedLinesParseScanner-16       1000000000    0.5926 ns/op   0 B/op   0 allocs/op
//
// BenchmarkShortLinesParseIndexAny-16      1000000000   0.06491 ns/op   0 B/op   0 allocs/op
// BenchmarkShortLinesParseReplaceCut-16    1000000000   0.06518 ns/op   0 B/op   0 allocs/op
// BenchmarkShortLinesParseCuts-16          1000000000   0.06749 ns/op   0 B/op   0 allocs/op
// BenchmarkShortLinesParseSlice-16         1000000000   0.07001 ns/op   0 B/op   0 allocs/op
// BenchmarkShortLinesParseReplaceSplit-16  1000000000   0.07158 ns/op   0 B/op   0 allocs/op
// BenchmarkShortLinesParseScanner-16       1000000000   0.18980 ns/op   0 B/op   0 allocs/op

var (
	count  = 100000
	prefix = "[prefix] "
	// dst = new(bytes.Buffer)
	// dst = io.Discard
	dst = os.Stdout

	shortLineBytes = bytes.Repeat(append(bytes.Repeat([]byte{'X'}, 100), bn[0]), count)
	longLineBytes  = bytes.Repeat(append(bytes.Repeat([]byte{'X'}, 10000), bn[0]), count)
	mixedLineBytes = bytes.Repeat([]byte("XXXXX\r\nXXXXX\n\n"), count)
)

func shortLines() io.Reader {
	return bytes.NewReader(shortLineBytes)
}

func longLines() io.Reader {
	return bytes.NewReader(longLineBytes)
}

func mixedLines() io.Reader {
	return bytes.NewReader(mixedLineBytes)
}

func BenchmarkShortLinesParseCuts(b *testing.B) {
	parseCuts(dst, shortLines(), prefix)
}

func BenchmarkLongLinesParseCuts(b *testing.B) {
	parseCuts(dst, longLines(), prefix)
}

func BenchmarkMixedLinesParseCuts(b *testing.B) {
	parseCuts(dst, mixedLines(), prefix)
}

func BenchmarkShortLinesParseIndexAny(b *testing.B) {
	parseIndexAny(dst, shortLines(), prefix)
}

func BenchmarkLongLinesParseIndexAny(b *testing.B) {
	parseIndexAny(dst, longLines(), prefix)
}

func BenchmarkMixedLinesParseIndexAny(b *testing.B) {
	parseIndexAny(dst, mixedLines(), prefix)
}

func BenchmarkShortLinesParseReplaceCut(b *testing.B) {
	parseReplaceCut(dst, shortLines(), prefix)
}

func BenchmarkLongLinesParseReplaceCut(b *testing.B) {
	parseReplaceCut(dst, longLines(), prefix)
}

func BenchmarkMixedLinesParseReplaceCut(b *testing.B) {
	parseReplaceCut(dst, mixedLines(), prefix)
}

func BenchmarkShortLinesParseSlice(b *testing.B) {
	parseSlice(dst, shortLines(), prefix)
}

func BenchmarkLongLinesParseSlice(b *testing.B) {
	parseSlice(dst, longLines(), prefix)
}

func BenchmarkMixedLinesParseSlice(b *testing.B) {
	parseSlice(dst, mixedLines(), prefix)
}

func BenchmarkShortLinesParseReplaceSplit(b *testing.B) {
	parseReplaceSplit(dst, shortLines(), prefix)
}

func BenchmarkLongLinesParseReplaceSplit(b *testing.B) {
	parseReplaceSplit(dst, longLines(), prefix)
}

func BenchmarkMixedLinesParseReplaceSplit(b *testing.B) {
	parseReplaceSplit(dst, mixedLines(), prefix)
}

func BenchmarkShortLinesParseScanner(b *testing.B) {
	parseScanner(dst, shortLines(), prefix)
}

func BenchmarkLongLinesParseScanner(b *testing.B) {
	parseScanner(dst, longLines(), prefix)
}

func BenchmarkMixedLinesParseScanner(b *testing.B) {
	parseScanner(dst, mixedLines(), prefix)
}

// r2n -prefix="[sh] " -stdio=stdout -- sh -c 'yes 1 | tr -d "\n" | head -c 100000'
// 과 같이 테스트 해봤는데 항상 4096B 에서 잘리는 것을 확인.
// stdin이 파이프라서 src.Read(buf)가 보통 4096B 단위로 처리되는 듯.
//
// 아래와 같이 Go로 처리하면 제약없이 테스트 가능.
// if out.Len() > maxLineLength 조건문이 없으면 out.Len()이 무한 증식하는 것이 확인된다.
func TestLongLine(t *testing.T) {
	data := bytes.Repeat([]byte{'X'}, 100000) // 100KB without '\r' or '\n'
	src := bytes.NewReader([]byte(data))
	prefix := "[prefix] "
	parseCuts(os.Stdout, src, prefix)
}

package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// Benchmark 결과:
//   최대한 표준 라이브러리를 쓰자. bytes.IndexByte(), bytes.Cuts() 성능이 평균적으로 가장 좋다.
//   slice 재할당은 성능과 메모리에 치명적이다.
//
// BenchmarkLongLinesParseIndexByte-16       1000000000       0.1824 ns/op       0 B/op   0 allocs/op
// BenchmarkLongLinesParseCuts-16            1000000000       0.1889 ns/op       0 B/op   0 allocs/op
// BenchmarkLongLinesParseReplaceCut-16      1000000000       0.2392 ns/op       1 B/op   0 allocs/op
// BenchmarkLongLinesParseReplaceSplit-16    1000000000       0.3350 ns/op       1 B/op   0 allocs/op
// BenchmarkLongLinesParseScanner-16         1000000000       0.9227 ns/op       0 B/op   0 allocs/op
// BenchmarkLongLinesParseIndexAny-16                 1   1051336208 ns/op   43120 B/op   7 allocs/op
// BenchmarkLongLinesParseSlice-16                    1   1535162542 ns/op   43120 B/op   7 allocs/op
//
// BenchmarkMixedLinesParseIndexAny-16       1000000000       0.1100 ns/op       0 B/op   0 allocs/op
// BenchmarkMixedLinesParseReplaceCut-16     1000000000       0.1106 ns/op       0 B/op   0 allocs/op
// BenchmarkMixedLinesParseSlice-16          1000000000       0.1134 ns/op       0 B/op   0 allocs/op
// BenchmarkMixedLinesParseIndexByte-16      1000000000       0.1164 ns/op       0 B/op   0 allocs/op
// BenchmarkMixedLinesParseCuts-16           1000000000       0.1229 ns/op       0 B/op   0 allocs/op
// BenchmarkMixedLinesParseReplaceSplit-16   1000000000       0.2407 ns/op       0 B/op   0 allocs/op
// BenchmarkMixedLinesParseScanner-16        1000000000       0.5082 ns/op       0 B/op   0 allocs/op
//
// BenchmarkShortLinesParseReplaceCut-16     1000000000      0.05670 ns/op       0 B/op   0 allocs/op
// BenchmarkShortLinesParseSlice-16          1000000000      0.05748 ns/op       0 B/op   0 allocs/op
// BenchmarkShortLinesParseCuts-16           1000000000      0.05758 ns/op       0 B/op   0 allocs/op
// BenchmarkShortLinesParseIndexByte-16      1000000000      0.05853 ns/op       0 B/op   0 allocs/op
// BenchmarkShortLinesParseIndexAny-16       1000000000      0.05902 ns/op       0 B/op   0 allocs/op
// BenchmarkShortLinesParseReplaceSplit-16   1000000000      0.06175 ns/op       0 B/op   0 allocs/op
// BenchmarkShortLinesParseScanner-16        1000000000      0.18540 ns/op       0 B/op   0 allocs/op

var (
	count  = 100000
	prefix = "[prefix] "
	// dst = new(bytes.Buffer)
	// dst = io.Discard
	dst = os.Stdout

	shortLineBytes = bytes.Repeat(append(bytes.Repeat([]byte{'X'}, 100), bsn[0]), count)
	longLineBytes  = bytes.Repeat(append(bytes.Repeat([]byte{'X'}, 10000), bsn[0]), count)
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

func BenchmarkShortLinesParseIndexByte(b *testing.B) {
	parseIndexByte(dst, shortLines(), prefix)
}

func BenchmarkLongLinesParseIndexByte(b *testing.B) {
	parseIndexByte(dst, longLines(), prefix)
}

func BenchmarkMixedLinesParseIndexByte(b *testing.B) {
	parseIndexByte(dst, mixedLines(), prefix)
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

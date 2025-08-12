package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// Benchmark 결과:
//   bytes.Cuts가 가장 빠르고 가벼우며, bytes.Replace 사용시에는 메모리가 급증한다.
//   bytes.IndexAny + 커스텀 로직 조합보다 그냥 최적화된 표준 라이브러리 쓰는게 낫다.
//   slice 재할당은 성능과 메모리에 치명적이다.
// BenchmarkLongLinesParseCuts-4                     1   1010934943 ns/op   1000170864 B/op   10 allocs/op
// BenchmarkLongLinesParseReplaceCut-4               1   1576265115 ns/op   3000379056 B/op   488343 allocs/op
// BenchmarkLongLinesParseReplaceSplit-4             1   1664624135 ns/op   3008639040 B/op   732509 allocs/op
// BenchmarkLongLinesParseScanner-4                  1   2348610999 ns/op   1000156496 B/op   9 allocs/op
// BenchmarkLongLinesParseIndexAny-4                 1   2742300694 ns/op   1000170864 B/op   10 allocs/op
// BenchmarkLongLinesParseReplaceSlice-4             1   3864148592 ns/op   4365047184 B/op   344174 allocs/op
// BenchmarkMixedLinesParseCuts-4           1000000000      0.09428 ns/op            0 B/op   0 allocs/op
// BenchmarkMixedLinesParseIndexAny-4       1000000000       0.1936 ns/op            0 B/op   0 allocs/op
// BenchmarkMixedLinesParseReplaceSlice-4   1000000000       0.2141 ns/op            0 B/op   0 allocs/op
// BenchmarkMixedLinesParseReplaceCut-4     1000000000       0.2193 ns/op            0 B/op   0 allocs/op
// BenchmarkMixedLinesParseReplaceSplit-4   1000000000       0.2406 ns/op            0 B/op   0 allocs/op
// BenchmarkMixedLinesParseScanner-4                 1   1145075522 ns/op      1405008 B/op   5 allocs/op
// BenchmarkShortLinesParseReplaceSplit-4   1000000000       0.1077 ns/op            0 B/op   0 allocs/op
// BenchmarkShortLinesParseReplaceCut-4     1000000000       0.1167 ns/op            0 B/op   0 allocs/op
// BenchmarkShortLinesParseCuts-4           1000000000       0.1228 ns/op            0 B/op   0 allocs/op
// BenchmarkShortLinesParseIndexAny-4       1000000000       0.1364 ns/op            0 B/op   0 allocs/op
// BenchmarkShortLinesParseReplaceSlice-4   1000000000       0.1499 ns/op            0 B/op   0 allocs/op
// BenchmarkShortLinesParseScanner-4        1000000000       0.3072 ns/op            0 B/op   0 allocs/op

var (
	count  = 100000
	prefix = "[prefix] "
	// dst = new(bytes.Buffer)
	// dst = io.Discard
	dst = os.Stdout
)

func shortLines() io.Reader {
	data := bytes.Repeat(append(bytes.Repeat([]byte{'X'}, 100), bn[0]), count)
	return bytes.NewReader(data)
}

func longLines() io.Reader {
	data := bytes.Repeat(append(bytes.Repeat([]byte{'X'}, 10000), bn[0]), count)
	return bytes.NewReader(data)
}

func mixedLines() io.Reader {
	data := bytes.Repeat([]byte("XXXXX\r\nXXXXX\n\n"), count)
	return bytes.NewReader(data)
}

func BenchmarkWarmUp(b *testing.B) {
	io.Copy(io.Discard, shortLines())
	io.Copy(io.Discard, longLines())
	io.Copy(io.Discard, mixedLines())
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

func BenchmarkShortLinesParseReplaceSlice(b *testing.B) {
	parseReplaceSlice(dst, shortLines(), prefix)
}

func BenchmarkLongLinesParseReplaceSlice(b *testing.B) {
	parseReplaceSlice(dst, longLines(), prefix)
}

func BenchmarkMixedLinesParseReplaceSlice(b *testing.B) {
	parseReplaceSlice(dst, mixedLines(), prefix)
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

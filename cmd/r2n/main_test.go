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
// BenchmarkLongLinesParseCuts-16            1000000000       0.2237 ns/op            0 B/op        0 allocs/op
// BenchmarkLongLinesParseReplaceCut-16      1000000000       0.3742 ns/op            2 B/op        0 allocs/op
// BenchmarkLongLinesParseReplaceSplit-16    1000000000       0.4269 ns/op            2 B/op        0 allocs/op
// BenchmarkLongLinesParseScanner-16         1000000000       0.9365 ns/op            0 B/op        0 allocs/op
// BenchmarkLongLinesParseIndexAny-16                 1   1073671583 ns/op        43120 B/op        7 allocs/op
// BenchmarkLongLinesParseReplaceSlice-16             1   1768642208 ns/op   3364902960 B/op   344168 allocs/op
// BenchmarkMixedLinesParseCuts-16           1000000000      0.06453 ns/op            0 B/op        0 allocs/op
// BenchmarkMixedLinesParseIndexAny-16       1000000000       0.1242 ns/op            0 B/op        0 allocs/op
// BenchmarkMixedLinesParseReplaceSlice-16   1000000000       0.1282 ns/op            0 B/op        0 allocs/op
// BenchmarkMixedLinesParseReplaceSplit-16   1000000000       0.1371 ns/op            0 B/op        0 allocs/op
// BenchmarkMixedLinesParseReplaceCut-16     1000000000       0.1421 ns/op            0 B/op        0 allocs/op
// BenchmarkMixedLinesParseScanner-16        1000000000       0.6581 ns/op            0 B/op        0 allocs/op
// BenchmarkShortLinesParseCuts-16           1000000000      0.06683 ns/op            0 B/op        0 allocs/op
// BenchmarkShortLinesParseIndexAny-16       1000000000      0.06889 ns/op            0 B/op        0 allocs/op
// BenchmarkShortLinesParseReplaceCut-16     1000000000      0.06972 ns/op            0 B/op        0 allocs/op
// BenchmarkShortLinesParseReplaceSplit-16   1000000000      0.06966 ns/op            0 B/op        0 allocs/op
// BenchmarkShortLinesParseReplaceSlice-16   1000000000      0.07796 ns/op            0 B/op        0 allocs/op
// BenchmarkShortLinesParseScanner-16        1000000000       0.1896 ns/op            0 B/op        0 allocs/op

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

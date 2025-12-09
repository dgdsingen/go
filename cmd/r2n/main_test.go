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
// BenchmarkLongLinesParseIndexByte-16      1000000000       0.1921 ns/op       0 B/op   0 allocs/op
// BenchmarkLongLinesParseCuts-16           1000000000       0.2048 ns/op       0 B/op   0 allocs/op
// BenchmarkLongLinesParseReplaceCut-16     1000000000       0.2686 ns/op       1 B/op   0 allocs/op
// BenchmarkLongLinesParseReplaceSplit-16   1000000000       0.3604 ns/op       1 B/op   0 allocs/op
// BenchmarkLongLinesParseScanner-16        1000000000       0.8052 ns/op       0 B/op   0 allocs/op
// BenchmarkLongLinesParseSlice-16                   1   1020488166 ns/op   43120 B/op   7 allocs/op
// BenchmarkLongLinesParseIndexAny-16                1   1079026541 ns/op   43120 B/op   7 allocs/op
//
// BenchmarkMixedLinesParseIndexByte-16     1000000000       0.1853 ns/op       0 B/op   0 allocs/op
// BenchmarkMixedLinesParseCuts-16          1000000000       0.2433 ns/op       0 B/op   0 allocs/op
// BenchmarkMixedLinesParseIndexAny-16      1000000000       0.2472 ns/op       0 B/op   0 allocs/op
// BenchmarkMixedLinesParseReplaceCut-16    1000000000       0.2522 ns/op       0 B/op   0 allocs/op
// BenchmarkMixedLinesParseSlice-16         1000000000       0.2571 ns/op       0 B/op   0 allocs/op
// BenchmarkMixedLinesParseScanner-16       1000000000       0.2598 ns/op       0 B/op   0 allocs/op
// BenchmarkMixedLinesParseReplaceSplit-16  1000000000       0.2682 ns/op       0 B/op   0 allocs/op
//
// BenchmarkShortLinesParseScanner-16       1000000000       0.06473 ns/op      0 B/op   0 allocs/op
// BenchmarkShortLinesParseIndexByte-16     1000000000       0.06546 ns/op      0 B/op   0 allocs/op
// BenchmarkShortLinesParseCuts-16          1000000000       0.06614 ns/op      0 B/op   0 allocs/op
// BenchmarkShortLinesParseReplaceSplit-16  1000000000       0.06618 ns/op      0 B/op   0 allocs/op
// BenchmarkShortLinesParseSlice-16         1000000000       0.06640 ns/op      0 B/op   0 allocs/op
// BenchmarkShortLinesParseReplaceCut-16    1000000000       0.06657 ns/op      0 B/op   0 allocs/op
// BenchmarkShortLinesParseIndexAny-16      1000000000       0.06727 ns/op      0 B/op   0 allocs/op

var (
	count = 100000
	// dst = bytes.Buffer{}
	// dst = io.Discard
	dst    = os.Stdout
	prefix = "[prefix] "

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
	parse(dst, shortLines(), &CutsParser{}, prefix)
}

func BenchmarkLongLinesParseCuts(b *testing.B) {
	parse(dst, longLines(), &CutsParser{}, prefix)
}

func BenchmarkMixedLinesParseCuts(b *testing.B) {
	parse(dst, mixedLines(), &CutsParser{}, prefix)
}

func BenchmarkShortLinesParseIndexByte(b *testing.B) {
	parse(dst, shortLines(), &IndexByteParser{}, prefix)
}

func BenchmarkLongLinesParseIndexByte(b *testing.B) {
	parse(dst, longLines(), &IndexByteParser{}, prefix)
}

func BenchmarkMixedLinesParseIndexByte(b *testing.B) {
	parse(dst, mixedLines(), &IndexByteParser{}, prefix)
}

func BenchmarkShortLinesParseIndexAny(b *testing.B) {
	parse(dst, shortLines(), &IndexAnyParser{}, prefix)
}

func BenchmarkLongLinesParseIndexAny(b *testing.B) {
	parse(dst, longLines(), &IndexAnyParser{}, prefix)
}

func BenchmarkMixedLinesParseIndexAny(b *testing.B) {
	parse(dst, mixedLines(), &IndexAnyParser{}, prefix)
}

func BenchmarkShortLinesParseSlice(b *testing.B) {
	parse(dst, shortLines(), &SliceParser{}, prefix)
}

func BenchmarkLongLinesParseSlice(b *testing.B) {
	parse(dst, longLines(), &SliceParser{}, prefix)
}

func BenchmarkMixedLinesParseSlice(b *testing.B) {
	parse(dst, mixedLines(), &SliceParser{}, prefix)
}

func BenchmarkShortLinesParseReplaceCut(b *testing.B) {
	parse(dst, shortLines(), &ReplaceCutParser{}, prefix)
}

func BenchmarkLongLinesParseReplaceCut(b *testing.B) {
	parse(dst, longLines(), &ReplaceCutParser{}, prefix)
}

func BenchmarkMixedLinesParseReplaceCut(b *testing.B) {
	parse(dst, mixedLines(), &ReplaceCutParser{}, prefix)
}

func BenchmarkShortLinesParseReplaceSplit(b *testing.B) {
	parse(dst, shortLines(), &ReplaceSplitParser{}, prefix)
}

func BenchmarkLongLinesParseReplaceSplit(b *testing.B) {
	parse(dst, longLines(), &ReplaceSplitParser{}, prefix)
}

func BenchmarkMixedLinesParseReplaceSplit(b *testing.B) {
	parse(dst, mixedLines(), &ReplaceSplitParser{}, prefix)
}

// func BenchmarkShortLinesParseScanner(b *testing.B) {
// 	parseScanner(dst, shortLines(), prefix)
// }
//
// func BenchmarkLongLinesParseScanner(b *testing.B) {
// 	parseScanner(dst, longLines(), prefix)
// }
//
// func BenchmarkMixedLinesParseScanner(b *testing.B) {
// 	parseScanner(dst, mixedLines(), prefix)
// }

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
	parse(os.Stdout, src, &IndexByteParser{}, prefix)
}

package main

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
)

// Benchmark 결과:
//   Replace 시리즈는 할당과 리소스 사용률이 너무 높아 의미가 없어서 순위 밖으로 빼버림.
//   성능/리소스 비율이 평균적으로 가장 좋은건 WindowParser > IndexByteParser 순.
//   최대한 표준 라이브러리를 쓰자. slice 재할당은 성능과 메모리에 치명적이다.
//
// BenchmarkParse/Short/Discard/IndexByte-18         	     883	   1223357 ns/op	6857.08 MB/s	  262128 B/op	      15 allocs/op
// BenchmarkParse/Short/Discard/Window-18            	     918	   1301177 ns/op	6446.98 MB/s	  262128 B/op	      15 allocs/op
// BenchmarkParse/Short/Discard/Slice-18             	     448	   2665683 ns/op	3146.91 MB/s	  262140 B/op	      15 allocs/op
// BenchmarkParse/Short/Discard/IndexAny-18          	     378	   3147047 ns/op	2665.56 MB/s	  262128 B/op	      15 allocs/op
// BenchmarkParse/Short/Discard/Cuts-18              	      82	  14305904 ns/op	 586.38 MB/s	  262128 B/op	      15 allocs/op
//
// BenchmarkParse/Short/DevNull/IndexByte-18         	     860	   1379860 ns/op	6079.35 MB/s	  262128 B/op	      15 allocs/op
// BenchmarkParse/Short/DevNull/Window-18            	     832	   1445842 ns/op	5801.92 MB/s	  262128 B/op	      15 allocs/op
// BenchmarkParse/Short/DevNull/Slice-18             	     421	   2843866 ns/op	2949.74 MB/s	  262128 B/op	      15 allocs/op
// BenchmarkParse/Short/DevNull/IndexAny-18          	     356	   3381264 ns/op	2480.92 MB/s	  262128 B/op	      15 allocs/op
// BenchmarkParse/Short/DevNull/Cuts-18              	      80	  14494112 ns/op	 578.76 MB/s	  262128 B/op	      15 allocs/op
//
// BenchmarkParse/Long/Discard/IndexByte-18          	    2342	    507847 ns/op	16522.37 MB/s	  202864 B/op	       8 allocs/op
// BenchmarkParse/Long/Discard/Window-18             	    1764	    677267 ns/op	12389.26 MB/s	  202864 B/op	       8 allocs/op
// BenchmarkParse/Long/Discard/Cuts-18               	    1898	    635243 ns/op	13208.86 MB/s	  202864 B/op	       8 allocs/op
// BenchmarkParse/Long/Discard/Slice-18              	     472	   2531394 ns/op	3314.71 MB/s	  202864 B/op	       8 allocs/op
// BenchmarkParse/Long/Discard/IndexAny-18           	     408	   2918350 ns/op	2875.20 MB/s	  202864 B/op	       8 allocs/op
//
// BenchmarkParse/Long/DevNull/IndexByte-18          	    1810	    656315 ns/op	12784.77 MB/s	  202864 B/op	       8 allocs/op
// BenchmarkParse/Long/DevNull/Cuts-18               	    1500	    787265 ns/op	10658.21 MB/s	  202864 B/op	       8 allocs/op
// BenchmarkParse/Long/DevNull/Window-18             	    1425	    839196 ns/op	9998.66 MB/s	  202864 B/op	       8 allocs/op
// BenchmarkParse/Long/DevNull/Slice-18              	     439	   2720808 ns/op	3083.95 MB/s	  202864 B/op	       8 allocs/op
// BenchmarkParse/Long/DevNull/IndexAny-18           	     385	   3112652 ns/op	2695.72 MB/s	  202864 B/op	       8 allocs/op
//
// BenchmarkParse/Mixed/Discard/Slice-18             	      76	  16186882 ns/op	 518.24 MB/s	  393200 B/op	      16 allocs/op
// BenchmarkParse/Mixed/Discard/IndexByte-18         	      57	  20941778 ns/op	 400.57 MB/s	  393200 B/op	      16 allocs/op
// BenchmarkParse/Mixed/Discard/Window-18            	      55	  21594908 ns/op	 388.45 MB/s	  393200 B/op	      16 allocs/op
// BenchmarkParse/Mixed/Discard/IndexAny-18          	      54	  22042782 ns/op	 380.56 MB/s	  393200 B/op	      16 allocs/op
// BenchmarkParse/Mixed/Discard/Cuts-18              	      32	  36389953 ns/op	 230.52 MB/s	  393200 B/op	      16 allocs/op
//
// BenchmarkParse/Mixed/DevNull/Slice-18             	      70	  16461166 ns/op	 509.60 MB/s	  393200 B/op	      16 allocs/op
// BenchmarkParse/Mixed/DevNull/IndexByte-18         	      56	  20857307 ns/op	 402.19 MB/s	  393200 B/op	      16 allocs/op
// BenchmarkParse/Mixed/DevNull/Window-18            	      55	  21781484 ns/op	 385.13 MB/s	  393200 B/op	      16 allocs/op
// BenchmarkParse/Mixed/DevNull/IndexAny-18          	      54	  22032726 ns/op	 380.73 MB/s	  393200 B/op	      16 allocs/op
// BenchmarkParse/Mixed/DevNull/Cuts-18              	      32	  37029266 ns/op	 226.54 MB/s	  393200 B/op	      16 allocs/op
//
// BenchmarkParse/Progress/Discard/Window-18         	     866	   1371319 ns/op	6117.21 MB/s	  262128 B/op	      15 allocs/op
// BenchmarkParse/Progress/Discard/Slice-18          	     369	   3201734 ns/op	2620.03 MB/s	  262128 B/op	      15 allocs/op
// BenchmarkParse/Progress/Discard/IndexAny-18       	     361	   3302895 ns/op	2539.78 MB/s	  262128 B/op	      15 allocs/op
// BenchmarkParse/Progress/Discard/IndexByte-18      	      66	  18215727 ns/op	 460.52 MB/s	  262128 B/op	      15 allocs/op
// BenchmarkParse/Progress/Discard/Cuts-18           	      64	  18463174 ns/op	 454.34 MB/s	  262128 B/op	      15 allocs/op
//
// BenchmarkParse/Progress/DevNull/Window-18         	     774	   1549360 ns/op	5414.26 MB/s	  262128 B/op	      15 allocs/op
// BenchmarkParse/Progress/DevNull/Slice-18          	     349	   3347417 ns/op	2506.00 MB/s	  262128 B/op	      15 allocs/op
// BenchmarkParse/Progress/DevNull/IndexAny-18       	     342	   3497332 ns/op	2398.58 MB/s	  262128 B/op	      15 allocs/op
// BenchmarkParse/Progress/DevNull/IndexByte-18      	      64	  18420201 ns/op	 455.40 MB/s	  262128 B/op	      15 allocs/op
// BenchmarkParse/Progress/DevNull/Cuts-18           	      63	  18799205 ns/op	 446.22 MB/s	  262128 B/op	      15 allocs/op
//
// BenchmarkParse/Short/Discard/ReplaceCut-18        	     772	   1559234 ns/op	5379.99 MB/s	 8650788 B/op	     272 allocs/op
// BenchmarkParse/Short/Discard/ReplaceSplit-18      	     577	   2071982 ns/op	4048.61 MB/s	10748020 B/op	     530 allocs/op
// BenchmarkParse/Short/DevNull/ReplaceCut-18        	     722	   1638887 ns/op	5118.51 MB/s	 8650787 B/op	     272 allocs/op
// BenchmarkParse/Short/DevNull/ReplaceSplit-18      	     543	   2196151 ns/op	3819.71 MB/s	10748020 B/op	     530 allocs/op
// BenchmarkParse/Long/Discard/ReplaceCut-18         	    1449	    800180 ns/op	10486.19 MB/s	 8593779 B/op	     265 allocs/op
// BenchmarkParse/Long/Discard/ReplaceSplit-18       	    1239	    969059 ns/op	8658.75 MB/s	 8620675 B/op	     523 allocs/op
// BenchmarkParse/Long/DevNull/ReplaceCut-18         	    1227	    933201 ns/op	8991.46 MB/s	 8593779 B/op	     265 allocs/op
// BenchmarkParse/Long/DevNull/ReplaceSplit-18       	    1077	   1161798 ns/op	7222.29 MB/s	 8620683 B/op	     523 allocs/op
// BenchmarkParse/Mixed/Discard/ReplaceCut-18        	      40	  29266342 ns/op	 286.63 MB/s	 8781824 B/op	     273 allocs/op
// BenchmarkParse/Mixed/Discard/ReplaceSplit-18      	      27	  40671429 ns/op	 206.25 MB/s	67502277 B/op	     531 allocs/op
// BenchmarkParse/Mixed/DevNull/ReplaceCut-18        	      39	  29505557 ns/op	 284.31 MB/s	 8781826 B/op	     273 allocs/op
// BenchmarkParse/Mixed/DevNull/ReplaceSplit-18      	      27	  41239674 ns/op	 203.41 MB/s	67502264 B/op	     531 allocs/op
// BenchmarkParse/Progress/Discard/ReplaceCut-18     	     517	   2342589 ns/op	3580.93 MB/s	 8650772 B/op	     272 allocs/op
// BenchmarkParse/Progress/Discard/ReplaceSplit-18   	     390	   3140024 ns/op	2671.52 MB/s	11272294 B/op	     530 allocs/op
// BenchmarkParse/Progress/DevNull/ReplaceCut-18     	     483	   2477353 ns/op	3386.13 MB/s	 8650770 B/op	     272 allocs/op
// BenchmarkParse/Progress/DevNull/ReplaceSplit-18   	     370	   3220954 ns/op	2604.40 MB/s	11272290 B/op	     530 allocs/op

const benchDataSize = 8 << 20 // 8MB

var (
	// 100B 라인. 일반적인 로그 한 줄
	shortLineBytes = repeatTo(append(bytes.Repeat([]byte{'X'}, 100), bn), benchDataSize)
	// 10KB 라인. IndexByte의 SIMD 이득이 최대가 되는 케이스
	longLineBytes = repeatTo(append(bytes.Repeat([]byte{'X'}, 10000), bn), benchDataSize)
	// '\r\n', '\n\n' 이 섞인 극단적으로 짧은 라인. 사실상 비정상 케이스
	mixedLineBytes = repeatTo([]byte("XXXXX\r\nXXXXX\n\n"), benchDataSize)
	// r2n의 실제 용도인 curl -# progress bar. 80B 라인이 '\r' 로만 끝남
	progressLineBytes = repeatTo(append(bytes.Repeat([]byte{'#'}, 79), br), benchDataSize)

	prefix = "[prefix] "
)

func repeatTo(unit []byte, size int) []byte {
	return bytes.Repeat(unit, size/len(unit)+1)
}

// Parser는 상태를 가질 수 있으므로(ReplaceSplitParser) 호출마다 새로 만든다.
// sameSpec=false 인 파서는 '\r\n' / '\n\r' 을 한 라인으로 합치지 않아서 출력이 다르다.
// 속도 참고용이지 그대로 교체할 수 있는 후보가 아니다.
var parsers = []struct {
	name     string
	new      func() Parser
	sameSpec bool
}{
	{"IndexByte", func() Parser { return &IndexByteParser{} }, true},
	{"Window", func() Parser { return &WindowParser{} }, true},
	{"Slice", func() Parser { return &SliceParser{} }, false},
	{"Cuts", func() Parser { return &CutsParser{} }, false},
	{"IndexAny", func() Parser { return &IndexAnyParser{} }, false},
	{"ReplaceCut", func() Parser { return &ReplaceCutParser{} }, false},
	{"ReplaceSplit", func() Parser { return &ReplaceSplitParser{} }, false},
}

var datasets = []struct {
	name string
	data []byte
}{
	{"Short", shortLineBytes},
	{"Long", longLineBytes},
	{"Mixed", mixedLineBytes},
	{"Progress", progressLineBytes},
}

// dst를 os.Stdout으로 두면 터미널 I/O가 시간을 다 먹어서 파서 간 차이가 묻힌다.
// Discard = 순수 파싱 비용, DevNull = write syscall 비용 포함.
var dsts = []struct {
	name string
	open func(testing.TB) io.Writer
}{
	{"Discard", func(testing.TB) io.Writer { return io.Discard }},
	{"DevNull", openDevNull},
}

func openDevNull(tb testing.TB) io.Writer {
	tb.Helper()
	f, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() {
		err = f.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	})
	return f
}

func BenchmarkParse(b *testing.B) {
	for _, ds := range datasets {
		for _, dst := range dsts {
			for _, p := range parsers {
				b.Run(ds.name+"/"+dst.name+"/"+p.name, func(b *testing.B) {
					w := dst.open(b)
					b.SetBytes(int64(len(ds.data)))
					b.ReportAllocs()
					for b.Loop() {
						parse(w, bytes.NewReader(ds.data), p.new(), prefix)
					}
				})
			}
		}
	}
}

// 프로덕션 파서의 출력 스펙을 고정하는 골든 테스트.
// parse()를 최적화할 때 동작이 바뀌지 않았는지 확인하는 기준이 된다.
// 입력은 Taskfile.yml 의 test:r2n 이 쓰는 것과 같다.
func TestParseOutput(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			// '\r\n' 과 '\n\r' 은 한 라인으로 합치고, 의도된 '\n\n' 은 빈 라인을 남긴다
			name: "mixed",
			in:   "\r1\r2\n3\n4\r5\r\n6\n\r7\r\r8\n\n9\n",
			want: "\n1\n2\n3\n4\n5\n6\n7\n\n8\n\n9\n",
		},
		{name: "empty", in: "", want: ""},
		{name: "cr only", in: "a\rb\rc\r", want: "a\nb\nc\n"},
		{name: "lf only", in: "a\nb\nc\n", want: "a\nb\nc\n"},
		{name: "crlf", in: "a\r\nb\r\n", want: "a\nb\n"},
		{name: "blank lines kept", in: "a\n\n\nb\n", want: "a\n\n\nb\n"},
		// '\n' 없이 끝나도 마지막 라인은 배출된다
		{name: "no trailing newline", in: "a\nb", want: "a\nb\n"},
	}

	for _, p := range parsers {
		if !p.sameSpec {
			continue
		}
		t.Run(p.name, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					for _, prefix := range []string{"", "[p] "} {
						var got bytes.Buffer
						parse(&got, bytes.NewReader([]byte(tt.in)), p.new(), prefix)

						want := tt.want
						if prefix != "" && want != "" {
							want = prefix + strings.ReplaceAll(strings.TrimSuffix(want, "\n"), "\n", "\n"+prefix) + "\n"
						}
						if got.String() != want {
							t.Errorf("prefix=%q\n got: %q\nwant: %q", prefix, got.String(), want)
						}
					}
				})
			}
		})
	}
}

// r2n -prefix="[sh] " -stdio=stdout -- sh -c 'yes 1 | tr -d "\n" | head -c 100000'
// 과 같이 테스트 해봤는데 항상 4096B 에서 잘리는 것을 확인.
// stdin이 파이프라서 src.Read(buf)가 보통 4096B 단위로 처리되는 듯.
//
// 아래와 같이 Go로 처리하면 제약없이 테스트 가능.
// stream.Len() > maxLineLength 조건문이 없으면 stream.Len()이 무한 증식하는 것이 확인된다.
func TestLongLine(t *testing.T) {
	const size = 1 << 20 // 1MB without '\r' or '\n'
	src := bytes.NewReader(bytes.Repeat([]byte{'X'}, size))

	var got bytes.Buffer
	parse(&got, src, &IndexByteParser{}, "")

	lines := bytes.Split(bytes.TrimSuffix(got.Bytes(), bsn), bsn)
	if len(lines) < 2 {
		t.Fatalf("구분자 없이 %dB가 들어왔는데 %d 라인으로 나옴. 강제 배출이 동작하지 않음", size, len(lines))
	}

	// 라인 길이 상한은 read buffer 크기에 따라 달라지므로 정확한 라인 수는 단언하지 않는다.
	// 버퍼가 무한 증식하지 않는다는 것만 확인하면 된다.
	limit := maxLineLength + readBufLength
	total := 0
	for i, line := range lines {
		if len(line) > limit {
			t.Errorf("line %d 길이 %d > 상한 %d", i, len(line), limit)
		}
		total += len(line)
	}
	if total != size {
		t.Errorf("입력 %dB 중 %dB만 출력됨", size, total)
	}
}

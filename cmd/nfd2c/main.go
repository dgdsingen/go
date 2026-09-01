// nfd2c: dir/file을 nfc로 정규화
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// 한글 음절 조합 상수: Unicode UAX #15, Hangul Syllable Composition
const (
	lBase, vBase, tBase = 0x1100, 0x1161, 0x11A7
	sBase               = 0xAC00
	lCount, vCount      = 19, 21
	tCount              = 28
	nCount              = vCount * tCount // 588
)

// 분해된 한글 자모(L+V[+T])를 완성형 음절로 결합하고, 기타 문자는 그대로 통과
func toNFC(s string) string {
	r := []rune(s)
	out := make([]rune, 0, len(r))
	for i := 0; i < len(r); i++ {
		l, v := r[i]-lBase, rune(-1)
		if l < 0 || l >= lCount || i+1 >= len(r) {
			out = append(out, r[i])
			continue
		}
		if v = r[i+1] - vBase; v < 0 || v >= vCount {
			out = append(out, r[i])
			continue
		}
		syl, t := sBase+(l*vCount+v)*tCount, rune(0)
		i++
		if i+1 < len(r) {
			if x := r[i+1] - tBase; x > 0 && x < tCount {
				t, i = x, i+1
			}
		}
		out = append(out, syl+t)
	}
	return string(out)
}

var (
	dryRun         = flag.Bool("n", false, "dry-run")
	quiet          = flag.Bool("q", false, "quiet")
	fixed, skipped int
)

// 정규화 비민감 파일시스템에서 두 경로가 같은 실체인지 봄
func sameEntry(a, b string) bool {
	fa, err1 := os.Lstat(a)
	fb, err2 := os.Lstat(b)
	return err1 == nil && err2 == nil && os.SameFile(fa, fb)
}

// temp를 경유한 2단계 rename 으로 APFS 정규화 이슈를 우회
func rename(src, dst string) error {
	tmp := dst + ".nfd2c-tmp"
	for i := 0; ; i++ {
		if _, err := os.Lstat(tmp); err != nil || sameEntry(tmp, src) {
			break
		}
		tmp = fmt.Sprintf("%s.nfd2c-tmp%d", dst, i)
	}
	if err := os.Rename(src, tmp); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

func fix(parent, name string) {
	nfc := toNFC(name)
	if nfc == name {
		return
	}
	src, dst := filepath.Join(parent, name), filepath.Join(parent, nfc)
	if _, err := os.Lstat(dst); err == nil && !sameEntry(src, dst) {
		fmt.Fprintf(os.Stderr, "충돌, 건너뜀: %s\n", dst)
		skipped++
		return
	}
	if !*quiet {
		prefix := ""
		if *dryRun {
			prefix = "[dry] "
		}
		fmt.Printf("%s%s > %s\n", prefix, src, dst)
	}
	if !*dryRun {
		if err := rename(src, dst); err != nil {
			fmt.Fprintf(os.Stderr, "실패 %s: %v\n", src, err)
			skipped++
			return
		}
	}
	fixed++
}

// 자식을 먼저 처리한 뒤 자신을 처리 (bottom-up)
func walk(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "읽기 실패 %s: %v\n", dir, err)
		skipped++
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			walk(filepath.Join(dir, e.Name())) // 심볼릭 링크는 IsDir=false
		}
		fix(dir, e.Name())
	}
}

func main() {
	flag.Parse()
	target := "."
	if flag.NArg() > 0 {
		target = flag.Arg(0)
	}
	root, err := filepath.Abs(strings.TrimRight(target, string(os.PathSeparator)))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	walk(root)
	fix(filepath.Dir(root), filepath.Base(root)) // root dir

	if !*quiet {
		fmt.Printf("\n%d개 정규화", fixed)
		if skipped > 0 {
			fmt.Printf(", %d개 건너뜀", skipped)
		}
		fmt.Println()
	}
	if skipped > 0 {
		os.Exit(1)
	}
}

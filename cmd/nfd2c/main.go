// nfd2c: dir/file을 nfc로 정규화
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/unicode/norm"
)

var (
	appName        = "nfd2c"
	version        = "undefined"
	versionFlag    = flag.Bool("version", false, "Version")
	dryRun         = flag.Bool("n", false, "dry-run")
	showAll        = flag.Bool("a", false, "show all")
	fixed, skipped int
)

func fmtVersion() string {
	return fmt.Sprintf("%s %s", appName, version)
}

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
	nfc := norm.NFC.String(name)
	if nfc == name {
		if *showAll {
			fmt.Printf("%s\n", filepath.Join(parent, name))
		}
		return
	}
	src, dst := filepath.Join(parent, name), filepath.Join(parent, nfc)
	if _, err := os.Lstat(dst); err == nil && !sameEntry(src, dst) {
		slog.Error(err.Error(), slog.String("dst", dst))
		skipped++
		return
	}
	prefix := ""
	if *dryRun {
		prefix = "[dry] "
	}
	fmt.Printf("%s%s > %s\n", prefix, src, dst)
	if !*dryRun {
		if err := rename(src, dst); err != nil {
			slog.Error(err.Error(), slog.String("src", src))
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
		slog.Error(err.Error(), slog.String("dir", dir))
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

	if *versionFlag {
		fmt.Println(fmtVersion())
		return
	}

	target := "."
	if flag.NArg() > 0 {
		target = flag.Arg(0)
	}
	root, err := filepath.Abs(strings.TrimRight(target, string(os.PathSeparator)))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(2)
	}
	fmt.Printf("nfd2c: %s\n", root)

	walk(root)
	fix(filepath.Dir(root), filepath.Base(root)) // root dir

	if fixed > 0 {
		fmt.Println()
	}
	fmt.Printf("Fix: %d", fixed)
	if skipped > 0 {
		fmt.Printf(", Skip: %d", skipped)
	}
	fmt.Println()
	if skipped > 0 {
		os.Exit(1)
	}
}

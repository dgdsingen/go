// nfd2c: dir/file을 nfc로 정규화
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/text/unicode/norm"
)

var (
	appName = "nfd2c"
	version = "undefined"

	tmpSuffix      = fmt.Sprintf(".%s-tmp", appName)
	fixed, skipped atomic.Int64

	versionFlag = flag.Bool("version", false, "version")
	dryRun      = flag.Bool("n", false, "dry-run")
	showAll     = flag.Bool("a", false, "show all")
	jobs        = flag.Int("j", runtime.NumCPU(), "jobs")
)

func fmtVersion() string {
	return fmt.Sprintf("%s %s", appName, version)
}

type WorkerPool struct {
	work chan func()
	wg   sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
	worker := &WorkerPool{work: make(chan func())}
	for range max(1, workers) {
		worker.wg.Go(func() {
			for f := range worker.work {
				f()
			}
		})
	}
	return worker
}

func (w *WorkerPool) Close() {
	close(w.work)
	w.wg.Wait()
}

// 정규화 비민감 파일시스템에서 두 경로가 같은 실체인지 확인
func sameEntry(a, b string) bool {
	fa, err1 := os.Lstat(a)
	fb, err2 := os.Lstat(b)
	return err1 == nil && err2 == nil && os.SameFile(fa, fb)
}

// temp를 경유한 2단계 rename 으로 APFS 정규화 이슈 우회
func rename(src, dst string) error {
	tmp := dst + tmpSuffix
	for i := 0; ; i++ {
		if _, err := os.Lstat(tmp); err != nil || sameEntry(tmp, src) {
			break
		}
		tmp = fmt.Sprintf("%s%s%d", dst, tmpSuffix, i)
	}
	if err := os.Rename(src, tmp); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

func fix(parent, name string, printWp *WorkerPool) {
	nfc := norm.NFC.String(name)
	if nfc == name {
		if *showAll {
			printWp.work <- func() {
				fmt.Println(filepath.Join(parent, name))
			}
		}
		return
	}
	src, dst := filepath.Join(parent, name), filepath.Join(parent, nfc)
	if _, err := os.Lstat(dst); err == nil && !sameEntry(src, dst) {
		slog.Error("dst already exists", slog.String("src", src), slog.String("dst", dst))
		skipped.Add(1)
		return
	}
	prefix := ""
	if *dryRun {
		prefix = "[dry] "
	}
	printWp.work <- func() {
		fmt.Printf("%s%s > %s\n", prefix, src, dst)
	}
	if !*dryRun {
		if err := rename(src, dst); err != nil {
			slog.Error(err.Error(), slog.String("src", src))
			skipped.Add(1)
			return
		}
	}
	fixed.Add(1)
}

// 자식 먼저 처리 후 자신을 처리 (bottom-up)
func walk(dir string, dirWp, fileWp, printWp *WorkerPool) *sync.WaitGroup {
	var wg sync.WaitGroup
	entries, err := os.ReadDir(dir)
	if err != nil {
		slog.Error(err.Error(), slog.String("dir", dir))
		skipped.Add(1)
		return &wg
	}
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() { // Symlink = !IsDir
			wg.Add(1)
			fileWp.work <- func() {
				defer wg.Done()
				fix(dir, name, printWp)
			}
			continue
		}
		sub := walk(filepath.Join(dir, name), dirWp, fileWp, printWp)
		// 하위 dir rename은 이미 큐에 먼저 들어갔으니, 남은 조건은 이 dir 바로 아래 file 뿐
		dirWp.work <- func() {
			sub.Wait()
			fix(dir, name, printWp)
		}
	}
	return &wg
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

	printWp := NewWorkerPool(1)
	dirWp := NewWorkerPool(1)
	fileWp := NewWorkerPool(*jobs)
	walk(root, dirWp, fileWp, printWp)
	dirWp.Close()
	fileWp.Close()
	printWp.Close()

	fmt.Printf("Fix: %d\n", fixed.Load())
	if n := skipped.Load(); n > 0 {
		fmt.Printf("Skip: %d\n", n)
		os.Exit(1)
	}
}

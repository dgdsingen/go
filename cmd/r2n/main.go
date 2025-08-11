package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var appName = "r2n"
var version = "undefined"

func fmtVersion() string {
	return fmt.Sprintf("%s %s", appName, version)
}

func main() {
	stdio := flag.String("stdio", "stderr", "stdio to replace [stdout, stderr, all]")
	prefix := flag.String("prefix", "", "prefix for each line")
	versionFlag := flag.Bool("version", false, "r2n version")
	flag.Parse()

	if *versionFlag {
		fmt.Println(fmtVersion())
		return
	}

	cmd := os.Args[0]
	args := flag.Args()
	if len(args) == 0 {
		usage := strings.Join([]string{
			fmtVersion(), "",
			"Usage:",
			"  " + cmd + " <command> [args...]",
			"  " + cmd + " -stdio=stdout -- <command> [args...]",
			"  " + cmd + ` -prefix="[curl] " -- <command> [args...]`,
			"",
		}, "\n")
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
		return
	}

	subCmd := exec.Command(args[0], args[1:]...)

	stdout, err := subCmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderr, err := subCmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	if err := subCmd.Start(); err != nil {
		panic(err)
	}

	switch *stdio {
	case "all":
		go copyAndReplace(os.Stdout, stdout, prefix)
		go copyAndReplace(os.Stderr, stderr, prefix)
	case "stdout":
		go copyAndReplace(os.Stdout, stdout, prefix)
		go io.Copy(os.Stderr, stderr)
	case "stderr":
		go io.Copy(os.Stdout, stdout)
		go copyAndReplace(os.Stderr, stderr, prefix)
	default:
		// stdout은 변환 없이 그대로 전달 (pipe 전달시 데이터 내용이 바뀌면 안됨)
		go io.Copy(os.Stdout, stdout)
		// stderr는 변환 후 전달 (curl의 progress bar 출력용)
		go copyAndReplace(os.Stderr, stderr, prefix)
	}

	if err := subCmd.Wait(); err != nil {
		os.Exit(subCmd.ProcessState.ExitCode())
	}
}

func replaceRN(bs []byte) []byte {
	p := 0
	prev := byte(0)
	for _, b := range bs {
		if b == '\r' {
			b = '\n'
		}
		if b == '\n' && prev == '\n' {
			continue
		}
		bs[p] = b
		p++
		prev = b
	}
	return bs[:p]
}

func writeAndFlushAll(dst *bufio.Writer, bs ...[]byte) {
	for _, b := range bs {
		dst.Write(b)
	}
	dst.Flush()
}

func copyAndReplace(dst io.Writer, src io.Reader, prefix *string) {
	const maxLineLength = 64 * 1024 // 64KB

	buf := make([]byte, 4096)
	// len > 0 이면 slice가 zero value로 채워져서 이상하게 출력될 수 있으므로 0으로 설정
	out := make([]byte, 0, 4096)
	// system call을 줄이기 위해 라인 단위로 버퍼링해서 출력
	dstBuf := bufio.NewWriter(dst)
	bprefix := []byte(*prefix)
	bn := []byte{'\n'}

	for {
		n, err := src.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			chunk = replaceRN(chunk)
			out = append(out, chunk...)

			// 예를 들어 "12\n34\n5" 중 "12", "34"는 각각의 라인으로 잘라서 전송하고
			p := 0
			for i, b := range out {
				if b == '\n' {
					writeAndFlushAll(dstBuf, bprefix, out[p:i+1])
					p = i + 1
				}
			}

			// 마지막 5는 아직 라인이 미완성이므로 버퍼에 남겨둠
			if p < len(out) {
				out = out[p:]
			} else {
				out = out[:0]
			}

			// chunk가 '\n' 없이 계속 들어올때 out 무한 증가를 막기 위해 강제 라인처리 + flush
			if len(out) > maxLineLength {
				writeAndFlushAll(dstBuf, bprefix, out, bn)
				out = out[:0]
			}
		}

		if err != nil {
			// '\n' 없이 끝난 경우 강제로 라인 처리해서 내보냄
			if len(out) > 0 {
				writeAndFlushAll(dstBuf, bprefix, out, bn)
			}
			break
		}
	}
}

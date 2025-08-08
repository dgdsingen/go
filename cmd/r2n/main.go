package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func main() {
	stdio := flag.String("stdio", "stderr", "Select stdio to replace [stdout, stderr]")
	flag.Parse()

	remainArgs := flag.Args()
	if len(remainArgs) < 1 {
		usage := strings.Join([]string{
			"Usage:",
			"  %s <command> [args...]",
			"  %s -stdio=stdout -- <command> [args...]",
			"\r",
		}, "\n")
		fmt.Fprintf(os.Stderr, usage, os.Args[0], os.Args[0])
		os.Exit(1)
		return
	}

	cmd := exec.Command(remainArgs[0], remainArgs[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	switch *stdio {
	case "stdout":
		go copyAndReplace(os.Stdout, stdout)
		go io.Copy(os.Stderr, stderr)
	case "stderr":
		go io.Copy(os.Stdout, stdout)
		go copyAndReplace(os.Stderr, stderr)
	default:
		// stdout은 변환 없이 그대로 전달 (pipe 전달시 데이터 내용이 바뀌면 안됨)
		go io.Copy(os.Stdout, stdout)
		// stderr는 변환 후 전달 (curl의 progress bar 출력용)
		go copyAndReplace(os.Stderr, stderr)
	}

	if err := cmd.Wait(); err != nil {
		os.Exit(cmd.ProcessState.ExitCode())
	}
}

func copyAndReplace(dst io.Writer, src io.Reader) {
	buf := make([]byte, 4096)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			// stderr \r > \n 변환
			out := strings.ReplaceAll(string(buf[:n]), "\r", "\n")
			out = strings.ReplaceAll(out, "\n\n", "\n")
			dst.Write([]byte(out))
		}
		if err != nil {
			break
		}
	}
}

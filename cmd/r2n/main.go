package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var version = "undefined"

func main() {
	stdio := flag.String("stdio", "stderr", "stdio to replace [stdout, stderr, all]")
	prefix := flag.String("prefix", "", "prefix for each line")
	showVersion := flag.Bool("version", false, "r2n version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version)
		return
	}

	remainArgs := flag.Args()
	if len(remainArgs) < 1 {
		usage := strings.Join([]string{
			"Usage:",
			`  ` + os.Args[0] + ` <command> [args...]`,
			`  ` + os.Args[0] + ` -stdio=stdout -- <command> [args...]`,
			`  ` + os.Args[0] + ` -prefix="[curl] " -- <command> [args...]`,
			"\r",
		}, "\n")
		fmt.Fprintf(os.Stderr, usage)
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

	if err := cmd.Wait(); err != nil {
		os.Exit(cmd.ProcessState.ExitCode())
	}
}

func copyAndReplace(dst io.Writer, src io.Reader, prefix *string) {
	buf := make([]byte, 4096)
	out := new(bytes.Buffer)
	bprefix := []byte(*prefix)
	br := []byte{'\r'}
	bn := []byte{'\n'}
	bnn := []byte{'\n', '\n'}
	for {
		n, err := src.Read(buf)
		if n > 0 {
			token := buf[:n]
			token = bytes.ReplaceAll(token, br, bn)
			token = bytes.ReplaceAll(token, bnn, bn)

			out.Write(token)

			if !bytes.Contains(token, bn) {
				continue
			}

			split := bytes.Split(out.Bytes(), bn)
			for i, s := range split {
				if i == len(split)-1 {
					out.Reset()
					out.Write(s)
					break
				}
				dst.Write(bprefix)
				dst.Write(s)
				dst.Write(bn)
			}
		}

		if err != nil {
			break
		}
	}
}

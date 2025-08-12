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
		go parseCuts(os.Stdout, stdout, *prefix)
		go parseCuts(os.Stderr, stderr, *prefix)
	case "stdout":
		go parseCuts(os.Stdout, stdout, *prefix)
		go io.Copy(os.Stderr, stderr)
	case "stderr":
		go io.Copy(os.Stdout, stdout)
		go parseCuts(os.Stderr, stderr, *prefix)
	default:
		// stdout은 변환 없이 그대로 전달 (pipe 전달시 데이터 내용이 바뀌면 안됨)
		go io.Copy(os.Stdout, stdout)
		// stderr는 변환 후 전달 (curl의 progress bar 출력용)
		go parseCuts(os.Stderr, stderr, *prefix)
	}

	if err := subCmd.Wait(); err != nil {
		os.Exit(subCmd.ProcessState.ExitCode())
	}
}

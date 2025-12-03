package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

const appName = "concurrent"

var version = "undefined"

func fmtVersion() string {
	return fmt.Sprintf("%s %s", appName, version)
}

func runCmd(line string, wg *sync.WaitGroup) {
	defer wg.Done()

	fields := strings.Fields(line)

	cmd := exec.Command(fields[0], fields[1:]...)
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

	cmdWg := &sync.WaitGroup{}
	cmdWg.Add(2)
	go func() {
		defer cmdWg.Done()
		_, err := io.Copy(os.Stdout, stdout)
		if err != nil && err != io.EOF {
			fmt.Printf("%v\n", err)
		}
	}()
	go func() {
		defer cmdWg.Done()
		_, err := io.Copy(os.Stderr, stderr)
		if err != nil && err != io.EOF {
			fmt.Printf("%v\n", err)
		}
	}()
	cmdWg.Wait()

	if err := cmd.Wait(); err != nil {
		os.Exit(cmd.ProcessState.ExitCode())
	}
}

func main() {
	cmd := flag.String("cmd", "", "Command")
	count := flag.Int("count", 1, "Count")
	versionFlag := flag.Bool("version", false, "Version")
	flag.Parse()

	if *versionFlag {
		fmt.Println(fmtVersion())
		return
	}

	remainArgs := flag.Args()
	var args []any = make([]any, len(remainArgs))
	for i, s := range remainArgs {
		args[i] = s
	}

	wg := &sync.WaitGroup{}
	var reader io.Reader = os.Stdin
	if *cmd != "" {
		reader = strings.NewReader(*cmd)
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		line = fmt.Sprintf(line, args...)

		for cnt := range *count {
			cntLine := strings.ReplaceAll(line, "{{.Count}}", strconv.Itoa(cnt))
			wg.Add(1)
			go runCmd(cntLine, wg)
		}
	}
	wg.Wait()

	if err := scanner.Err(); err != nil {
		fmt.Printf("%v\n", err)
	}
}

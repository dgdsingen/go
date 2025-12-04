package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const appName = "concurrent"

var (
	version           = "undefined"
	quotePrefixRegexp = regexp.MustCompile(`^("|')(.*)`)
	quoteSuffixRegexp = regexp.MustCompile(`(.*)("|')$`)
)

func fmtVersion() string {
	return fmt.Sprintf("%s %s", appName, version)
}

// `ls -l "/some/spaces in path"` 를 ["ls", "-l", "/some/spaces in path"] 로 변경
func splitCmd(cmd string) (cmdSlice []string) {
	sb := &strings.Builder{}
	for s := range strings.FieldsSeq(cmd) {
		// "" or '' 시작 조건
		if quotePrefixRegexp.MatchString(s) {
			s = quotePrefixRegexp.FindStringSubmatch(s)[2]
			sb.WriteString(s)
			sb.WriteString(" ")
			continue
		}
		// "" or '' 종료 조건
		if quoteSuffixRegexp.MatchString(s) {
			s = quoteSuffixRegexp.FindStringSubmatch(s)[1]
			sb.WriteString(s)
			cmdSlice = append(cmdSlice, sb.String())
			sb.Reset()
			continue
		}
		// sb 내용물이 있으면 "" or '' 시작~종료 사이 상태이므로 sb에 넣고 아니면 cmdSlice에 넣음
		if sb.Len() == 0 {
			cmdSlice = append(cmdSlice, s)
		} else {
			sb.WriteString(s)
			sb.WriteString(" ")
		}
	}
	return cmdSlice
}

func addCmdArgs(cmd string, args []string) string {
	if cmd == "" {
		return strings.Join(args, " ")
	}
	for _, arg := range args {
		// 공백이 포함된 arg는 ""로 묶어줌
		if strings.Contains(arg, " ") {
			arg = `"` + arg + `"`
		}
		cmd = strings.Replace(cmd, "{}", arg, 1)
	}
	return cmd
}

func addCmdCount(cmd string, count int) string {
	return strings.ReplaceAll(cmd, "{{.Count}}", strconv.Itoa(count))
}

func runCmd(lines []string, wg *sync.WaitGroup) {
	defer wg.Done()

	cmd := exec.Command(lines[0], lines[1:]...)
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
	useStdin := flag.Bool("i", false, "Use stdin")
	versionFlag := flag.Bool("version", false, "Version")
	flag.Parse()

	if *versionFlag {
		fmt.Println(fmtVersion())
		return
	}

	// 적용 우선순위: -cmd > flag.Args() > Stdin
	// 예를 들어 `echo 3 | concurrent -i -cmd="echo 1 {} {}" 2` 실행시 "1 2 3" 출력됨
	flagArgsCmd := addCmdArgs(*cmd, flag.Args())
	if flagArgsCmd == "" && !*useStdin {
		fmt.Println("use -cmd or -i(stdin)")
		os.Exit(1)
	}

	wg := &sync.WaitGroup{}
	if *useStdin {
		// `echo 1 | concurrent -cmd="echo 1"` 실행시 정상이지만
		// `concurrent -cmd="echo 1"` 실행시 stdin 값이 들어올때까지 기다린다.
		// 이를 처리할 방법이 없고 직접 구현하면 너무 복잡할듯해 -i 옵션으로 처리.
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}

			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			for cnt := range *count {
				stdinArgsCmd := addCmdArgs(flagArgsCmd, []string{line})
				countCmd := addCmdCount(stdinArgsCmd, cnt)
				cmdSlice := splitCmd(countCmd)
				wg.Add(1)
				go runCmd(cmdSlice, wg)
			}
		}
	} else {
		for cnt := range *count {
			countCmd := addCmdCount(flagArgsCmd, cnt)
			cmdSlice := splitCmd(countCmd)
			wg.Add(1)
			go runCmd(cmdSlice, wg)
		}
	}
	wg.Wait()
}

/* Test
mkdir -p concurrent-test
touch "concurrent-test/1.txt"
touch "concurrent-test/some space.txt"

concurrent -cmd="echo 1"
concurrent -cmd="echo {}"
concurrent -cmd="echo {}" 1
concurrent -cmd="echo {}" 1 2
concurrent -cmd="echo {} {}" 1 2
concurrent -cmd="echo {} {}" 1 2 3

concurrent -cmd="echo 1" -i # 무한 대기하는게 정상
echo 3 | concurrent -cmd="echo 1" -i
echo 3 | concurrent -cmd="echo {}" -i
echo 3 | concurrent -cmd="echo {}" 1 -i
seq 3 | concurrent -cmd="echo {}" 1 -i
seq 3 | concurrent -cmd="echo {}" -i

concurrent -cmd="echo 1" -count=2
concurrent -cmd="echo {}" -count=2
concurrent -cmd="echo {}" -count=2 1

fd . concurrent-test | concurrent -cmd="ls -l {}" -i
concurrent -cmd='ls -l {}' "concurrent-test/some space.txt"

rm -rf concurrent-test
*/

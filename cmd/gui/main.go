package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-vgo/robotgo"
)

var (
	appName = "gui"
	version = "undefined"
)

func fmtVersion() string {
	return fmt.Sprintf("%s %s", appName, version)
}

func SecToDuration(s int) time.Duration {
	return time.Duration(s) * time.Second
}

func StepSec() int {
	return rand.Intn(5) + 20
}

func RandPointGen(points []int) func() int {
	return func() int {
		return points[rand.Intn(len(points))]
	}
}

// func hasProcess(procName string) bool {
// 	processes, err := process.Processes()
// 	if err != nil {
// 		fmt.Printf("%v\n", err)
// 	}
//
// 	for _, proc := range processes {
// 		name, err := proc.Name()
// 		if err != nil {
// 			// fmt.Printf("%v\n", err)
// 			continue
// 		}
// 		if proc.Pid != int32(pid) && name == procName {
// 			return true
// 		}
// 	}
//
// 	return false
// }

func PidFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return home + "/.gui.pid"
}

func existsPidFile(pidFilePath string) bool {
	if _, err := os.Stat(pidFilePath); !os.IsNotExist(err) {
		return true
	}
	return false
}

func readPidFile(pidFilePath string) (pid int) {
	if !existsPidFile(pidFilePath) {
		return -1
	}

	data, err := os.ReadFile(pidFilePath)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	pid, err = strconv.Atoi(string(data))
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return pid
}

func writePidFile(pidFilePath string, pid int) {
	err := os.WriteFile(pidFilePath, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

func deletePidFile(pidFilePath string) {
	err := os.Remove(pidFilePath)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

func Process(pid int) *os.Process {
	proc, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}
	return proc
}

func existsProcess(proc *os.Process) bool {
	return proc.Signal(syscall.Signal(0)) == nil
}

func exitProcess(pidFilePath string) {
	deletePidFile(pidFilePath)
	os.Exit(0)
}

func main() {
	on := flag.Bool("on", false, "on")
	totalSec := flag.Int("total-sec", -1, "total seconds")
	versionFlag := flag.Bool("version", false, "gui version")
	flag.Parse()

	if *versionFlag {
		fmt.Println(fmtVersion())
		return
	}

	// `gui -total-sec=3 on` 순서로 실행해야 정상 (args=[on])
	// `gui on -total-sec=3` 순서로 실행하면 안됨 (args=[on -total-sec=3])
	args := flag.Args()
	onoff := ""
	if len(args) == 1 {
		onoff = args[0]
	}

	pidFilePath := PidFilePath()
	pid := readPidFile(pidFilePath)
	proc := Process(pid)
	existsProc := existsProcess(proc)

	if existsProc && (*on || onoff == "on") {
		fmt.Printf("gui (PID=%d) is already running.\n", pid)
		os.Exit(0)
	}

	// `gui on` = background로 `gui -on` 띄우고 자신은 종료
	// `gui -on` = foreground로 gui를 실제 실행
	if !*on {
		switch onoff {
		case "on":
			cmd := exec.Command("gui", "-on", "-total-sec", strconv.Itoa(*totalSec))
			err := cmd.Start()
			if err != nil {
				fmt.Printf("%v\n", err)
			}
		case "off":
			err := proc.Signal(syscall.SIGTERM)
			if err != nil {
				fmt.Printf("%v\n", err)
			}
		default:
			fmt.Printf("gui (PID=%d) (exists=%v).\n", pid, existsProc)
		}
		os.Exit(0)
	}

	pid = os.Getpid()
	writePidFile(pidFilePath, pid)

	randPoint := RandPointGen([]int{-1, 1})

	ticker := time.NewTicker(SecToDuration(StepSec()))

	timer := &time.Timer{}
	if *totalSec > 0 {
		timer = time.NewTimer(SecToDuration(*totalSec))
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			robotgo.MoveRelative(randPoint(), randPoint())
			ticker.Reset(SecToDuration(StepSec()))
		case <-timer.C:
			sigChan <- syscall.SIGTERM
		case <-sigChan:
			exitProcess(pidFilePath)
		}
	}
}

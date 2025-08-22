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
	appName     = "gui"
	version     = "undefined"
	pid         = os.Getpid()
	pidFilePath = PidFilePath()
)

func fmtVersion() string {
	return fmt.Sprintf("%s %s", appName, version)
}

func StepSec() int {
	return rand.Intn(5) + 10
}

// func hasProcess(procName string) bool {
// 	processes, err := process.Processes()
// 	if err != nil {
// 		fmt.Printf("error processes: %v\n", err)
// 	}
//
// 	for _, proc := range processes {
// 		name, err := proc.Name()
// 		if err != nil {
// 			// fmt.Printf("error proc cmd: %v %v\n", cmd, err)
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

func existsPidFile() bool {
	if _, err := os.Stat(pidFilePath); !os.IsNotExist(err) {
		return true
	}
	return false
}

func readPidFile() (pid int) {
	if !existsPidFile() {
		return -1
	}

	data, err := os.ReadFile(pidFilePath)
	if err != nil {
		fmt.Printf("os.ReadFile(): %v\n", err)
	}
	pid, err = strconv.Atoi(string(data))
	if err != nil {
		fmt.Printf("Atoi(): %v\n", err)
	}
	return pid
}

func writePidFile() {
	os.WriteFile(pidFilePath, []byte(strconv.Itoa(pid)), 0644)
}

func deletePidFile() {
	os.Remove(pidFilePath)
}

func Process(pid int) (*os.Process, error) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("find proc error: %v\n", err)
		return nil, err
	}
	return proc, err
}

func existsProcess(proc *os.Process) bool {
	return proc.Signal(syscall.Signal(0)) == nil
}

func exitProcess() {
	deletePidFile()
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

	args := flag.Args()
	onoff := ""
	if len(args) == 1 {
		onoff = args[0]
	}

	pid := readPidFile()
	proc, err := Process(pid)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	// `gui on` = background로 `gui -on` 띄우고 자신은 종료
	switch onoff {
	case "on":
		if existsProcess(proc) {
			fmt.Printf("gui (PID=%d) is already running.\n", pid)
		} else {
			cmd := exec.Command("gui", "-on", "-total-sec", strconv.Itoa(*totalSec))
			err := cmd.Start()
			if err != nil {
				fmt.Printf("cmd error: %v\n", err)
			}
		}
	case "off":
		proc.Signal(syscall.SIGTERM)
	default:
		fmt.Printf("gui (PID=%d) (exists=%v).\n", pid, existsProcess(proc))
	}

	// `gui -on` = foreground로 gui를 실제 실행
	if !*on {
		os.Exit(0)
	}

	writePidFile()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	stepSec := StepSec()

	for {
		select {
		case <-done:
			exitProcess()
		default:
			time.Sleep(1 * time.Second)

			if *totalSec--; *totalSec == 0 {
				exitProcess()
			}

			if stepSec--; stepSec <= 0 {
				x := []int{-1, 1}[rand.Intn(2)]
				y := []int{-1, 1}[rand.Intn(2)]
				robotgo.MoveRelative(x, y)

				stepSec = StepSec()
			}
		}
	}
}

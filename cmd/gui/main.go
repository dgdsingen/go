package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-vgo/robotgo"
)

var (
	appName = "gui"
	version = "undefined"
	pid     = os.Getpid()
	pidFile = PidFile()
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

func PidFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return home + "/.gui.pid"
}

func readPidFile() []byte {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return data
}

func writePidFile() {
	os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
}

func deletePidFile() {
	os.Remove(pidFile)
}

func existsPidFile() bool {
	if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
		return true
	}
	return false
}

func main() {
	totalSec := flag.Int("total-sec", -1, "total seconds")
	versionFlag := flag.Bool("version", false, "gui version")
	flag.Parse()

	if *versionFlag {
		fmt.Println(fmtVersion())
		return
	}

	if existsPidFile() {
		fmt.Printf("gui (PID=%s) is already running.\n", readPidFile())
		os.Exit(1)
	}

	writePidFile()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	stepSec := StepSec()

	for {
		select {
		case <-done:
			deletePidFile()
			os.Exit(0)
		default:
			time.Sleep(1 * time.Second)

			if *totalSec--; *totalSec == 0 {
				break
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

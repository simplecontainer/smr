package agent

import (
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"os"
	"strconv"
	"strings"
)

func Stop() {
	StopFlannel()
	StopControl()
}

func StopFlannel() {
	err := StopProcessFromPIDFile("/var/run/flannel.pid")

	if err != nil {
		helpers.PrintAndExit(err, 1)
	} else {
		fmt.Println("process killed successfully")
	}
}

func StopControl() {
	err := StopProcessFromPIDFile("/var/run/control.pid")

	if err != nil {
		helpers.PrintAndExit(err, 1)
	} else {
		fmt.Println("process killed successfully")
	}
}

func StopProcessFromPIDFile(pidFile string) error {
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("failed to read pid file: %w", err)
	}

	pidStr := strings.TrimSpace(string(pidData))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return fmt.Errorf("invalid pid in file: %w", err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := proc.Kill(); err != nil {
		return fmt.Errorf("failed to kill process %d: %w", pid, err)
	}

	fmt.Printf("process %d killed successfully\n", pid)
	return nil
}

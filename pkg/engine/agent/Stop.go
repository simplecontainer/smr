package agent

import (
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"os"
	"strconv"
	"strings"
)

func StopFlannel() {
	pidFile := "/var/run/flannel.pid"
	err := StopProcessFromPIDFile(pidFile)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}
}

func StopControl() {
	pidFile := fmt.Sprintf("/var/run/user/%d/control.pid", os.Getuid())

	err := StopProcessFromPIDFile(pidFile)

	if err != nil {
		helpers.PrintAndExit(err, 1)
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

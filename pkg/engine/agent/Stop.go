package agent

import (
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"os"
	"strconv"
)

func Stop() {
	pidStr, err := os.ReadFile("/var/run/flannel.pid")

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	var pid int
	pid, err = strconv.Atoi(string(pidStr))

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	var proc *os.Process
	proc, err = os.FindProcess(pid)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	err = proc.Kill()

	if err != nil {
		helpers.PrintAndExit(err, 1)
	} else {
		fmt.Println("process killed successfully")
	}
}

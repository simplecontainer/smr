package manager

import (
	"fmt"
	"strings"
)

func (mgr *Manager) Info() {
	mgr.Generate()

	fmt.Println("Selected config")
	fmt.Println(strings.Repeat("-", 32))
	fmt.Println(fmt.Sprintf("Environment: %s", mgr.Config.Configuration.Environment.Target))
	fmt.Println(fmt.Sprintf("Project root: %s", mgr.Config.Configuration.Environment.Root))
}

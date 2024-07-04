package commands

import (
	"github.com/simplecontainer/smr/pkg/bootstrap"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"os"
)

func Create() {
	Commands = append(Commands, Command{
		name: "create",
		condition: func(*manager.Manager) bool {
			if os.Args[2] == "" {
				logger.Log.Warn("please specify project name")
				return false
			} else {
				return true
			}
		},
		functions: []func(*manager.Manager, []string){
			func(mgr *manager.Manager, args []string) {
				logger.Log.Info("created new project")
			},
		},
		depends_on: []func(*manager.Manager, []string){
			func(mgr *manager.Manager, args []string) {
				bootstrap.CreateProject(args[2], mgr.Config)
			},
		},
	})
}

package commands

import (
	"os"
	"smr/pkg/logger"
	"smr/pkg/manager"
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
				mgr.CreateProject(args[2])
			},
		},
	})
}

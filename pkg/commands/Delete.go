package commands

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/bootstrap"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/utils"
	"os"
)

func Delete() {
	Commands = append(Commands, Command{
		name: "delete",
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
				if utils.Confirm(fmt.Sprintf("Are you sure? Delete project %s is irreversible?", mgr.Config.Environment.PROJECT)) {
					bootstrap.DeleteProject(os.Args[2], mgr.Config)
				}

				os.Exit(0)
			},
		},
		depends_on: []func(*manager.Manager, []string){
			func(mgr *manager.Manager, args []string) {},
		},
	})
}

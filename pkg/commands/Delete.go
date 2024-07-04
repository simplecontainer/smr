package commands

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/bootstrap"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/utils"
	"github.com/spf13/viper"
)

func Delete() {
	Commands = append(Commands, Command{
		name: "delete",
		condition: func(*manager.Manager) bool {
			return viper.GetString("project") != ""
		},
		functions: []func(*manager.Manager, []string){
			func(mgr *manager.Manager, args []string) {
				if utils.Confirm(fmt.Sprintf("Are you sure? Delete project %s is irreversible?", mgr.Config.Environment.PROJECT)) {
					bootstrap.DeleteProject(viper.GetString("project"), mgr.Config)
				}
			},
		},
		depends_on: []func(*manager.Manager, []string){
			func(mgr *manager.Manager, args []string) {

			},
		},
	})
}

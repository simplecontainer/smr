package commands

import (
	"fmt"
	"smr/pkg/manager"
	"smr/pkg/utils"

	"github.com/spf13/viper"
)

func Down() {
	Commands = append(Commands, Command{
		name: "down",
		condition: func(*manager.Manager) bool {
			return viper.GetBool("down")
		},
		functions: []func(*manager.Manager, []string){
			func(mgr *manager.Manager, args []string) {
				if utils.Confirm(fmt.Sprintf("Do you want to bring down project %s?", mgr.Runtime.PROJECT)) {
					mgr.Generate()
				}
			},
		},
		depends_on: []func(*manager.Manager, []string){
			func(mgr *manager.Manager, args []string) {
				mgr.Config.Load(mgr.Runtime.PROJECTDIR)
			},
		},
	})
}

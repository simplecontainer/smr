package commands

import (
	"smr/pkg/manager"
)

func Ps() {
	Commands = append(Commands, Command{
		name:      "ps",
		condition: func(*manager.Manager) bool { return true },
		functions: []func(*manager.Manager, []string){
			func(mgr *manager.Manager, args []string) {
				mgr.OutputTable()
			},
		},
		depends_on: []func(*manager.Manager, []string){
			func(mgr *manager.Manager, args []string) {
				mgr.Config.Load(mgr.Runtime.PROJECTDIR)
			},
		},
	})
}

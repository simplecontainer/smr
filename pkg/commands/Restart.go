package commands

import (
	"smr/pkg/logger"
	"smr/pkg/manager"
)

func Restart() {
	Commands = append(Commands, Command{
		name: "restart",
		condition: func(*manager.Manager) bool {
			return true
		},
		functions: []func(*manager.Manager, []string){
			func(mgr *manager.Manager, args []string) {
				logger.Log.Info("Restarting container correctly")
			},
		},
		depends_on: []func(*manager.Manager, []string){
			func(mgr *manager.Manager, args []string) {},
		},
	})
}

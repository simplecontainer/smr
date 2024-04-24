package commands

import (
	"os"
	"smr/pkg/definitions"
	"smr/pkg/manager"
)

func Apply() {
	Commands = append(Commands, Command{
		name: "apply",
		condition: func(*manager.Manager) bool {
			return os.Args[2] != ""
		},
		functions: []func(*manager.Manager, []string){
			func(mgr *manager.Manager, args []string) {
				definition := definitions.ReadFile(args[2])
				mgr.Apply(definition)
			},
		},
		depends_on: []func(*manager.Manager, []string){
			func(mgr *manager.Manager, args []string) {
				mgr.Config.Load(mgr.Runtime.PROJECTDIR)
			},
		},
	})
}

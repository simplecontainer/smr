package commands

import (
	"os"
	"smr/pkg/manager"
)

var Commands []Command

func PreloadCommands() {
	Apply()
	Create()
	Delete()
	Ps()
}

func Run(mgr *manager.Manager) {
	for _, comm := range Commands {
		for _, arg := range os.Args {
			if comm.name == arg {
				if comm.condition(mgr) {
					for _, fn := range comm.depends_on {
						fn(mgr, os.Args)
					}

					for _, fn := range comm.functions {
						fn(mgr, os.Args)
					}
				}
			}
		}
	}
}

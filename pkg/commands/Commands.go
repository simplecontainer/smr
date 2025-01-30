package commands

import (
	"github.com/simplecontainer/smr/pkg/api"
	"os"
)

var Commands []Command

func PreloadCommands() {
	Create()
	Start()
}

func Run(api *api.Api) {
	for _, comm := range Commands {
		for _, arg := range os.Args {
			if comm.name == arg {
				if comm.condition(api) {
					for _, fn := range comm.depends_on {
						fn(api, os.Args)
					}

					for _, fn := range comm.functions {
						fn(api, os.Args)
					}
				}
			}
		}
	}
}

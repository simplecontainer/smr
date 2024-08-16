package commands

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/simplecontainer/smr/pkg/bootstrap"
	"github.com/simplecontainer/smr/pkg/helpers"
	"github.com/simplecontainer/smr/pkg/logger"
	"os"
)

func Delete() {
	Commands = append(Commands, Command{
		name: "delete",
		condition: func(*api.Api) bool {
			if os.Args[2] == "" {
				logger.Log.Warn("please specify project name")
				return false
			} else {
				return true
			}
		},
		functions: []func(*api.Api, []string){
			func(api *api.Api, args []string) {
				if helpers.Confirm(fmt.Sprintf("Are you sure? Delete project %s is irreversible?", api.Config.Environment.PROJECT)) {
					bootstrap.DeleteProject(os.Args[2], api.Config)
				}

				os.Exit(0)
			},
		},
		depends_on: []func(*api.Api, []string){
			func(api *api.Api, args []string) {},
		},
	})
}

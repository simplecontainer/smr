package commands

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/bootstrap"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	"io"
	"os"
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
				_, err := bootstrap.CreateProject(os.Args[2], mgr.Config)

				if err != nil {
					panic(err)
				}

				var out io.Writer
				out, err = os.Open(fmt.Sprintf("%s/%s/%s", mgr.Config.Environment.HOMEDIR, static.SMR, os.Args[2]))

				if err != nil {
					panic(err)
				}

				err = startup.Save(mgr.Config, out)

				if err != nil {
					panic(err)
				}
			},
		},
		depends_on: []func(*manager.Manager, []string){
			func(mgr *manager.Manager, args []string) {},
		},
	})
}

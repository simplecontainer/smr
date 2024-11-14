package commands

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/simplecontainer/smr/pkg/bootstrap"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/viper"
	"io"
	"os"
	"strings"
)

func Create() {
	Commands = append(Commands, Command{
		name: "create",
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
				_, err := bootstrap.CreateProject(os.Args[2], api.Config)

				if err != nil {
					panic(err)
				}

				var out io.Writer
				out, err = os.OpenFile(fmt.Sprintf("%s/%s/%s/%s/config.yaml", api.Config.Environment.HOMEDIR, static.SMR, os.Args[2], static.CONFIGDIR), (os.O_WRONLY | os.O_CREATE), 0644)

				if err != nil {
					panic(err)
				}

				// TODO: Translate all these to viper flags it is ugly

				target := "development"
				if os.Getenv("ENVIRONMENT") != "" {
					target = os.Getenv("ENVIRONMENT")
				}

				domain := "localhost"
				if os.Getenv("DOMAIN") != "" {
					domain = os.Getenv("DOMAIN")
				}

				externalIP := "127.0.0.1"
				if os.Getenv("EXTERNALIP") != "" {
					externalIP = os.Getenv("EXTERNALIP")
				}

				hostHomeDir := ""
				if os.Getenv("HOMEDIR") != "" {
					hostHomeDir = os.Getenv("HOMEDIR")
				}

				platform := static.PLATFORM_DOCKER
				if os.Getenv("PLATFORM") != "" {
					platform = os.Getenv("PLATFORM")
				}

				hostname := ""
				if os.Getenv("HOSTNAME") != "" {
					hostname = os.Getenv("HOSTNAME")
				}

				api.Config.Platform = platform
				api.Config.OverlayNetwork = viper.GetString("overlay")
				api.Config.Port = viper.GetInt("port")
				api.Config.Agent = viper.GetString("agent")
				api.Config.Target = target
				api.Config.Root = api.Config.Environment.PROJECTDIR
				api.Config.Domain = domain
				api.Config.ExternalIP = externalIP
				api.Config.OptRoot = "/opt/smr"
				api.Config.CommonName = "root"
				api.Config.HostHome = hostHomeDir
				api.Config.Node = hostname

				api.Config.KVStore = &configuration.KVStore{
					Cluster:     strings.Split(viper.GetString("cluster"), ","),
					Node:        uint64(viper.GetInt("node")),
					URL:         viper.GetString("url"),
					JoinCluster: viper.GetBool("join"),
				}

				err = startup.Save(api.Config, out)

				if err != nil {
					panic(err)
				}
			},
		},
		depends_on: []func(*api.Api, []string){
			func(api *api.Api, args []string) {},
		},
	})
}

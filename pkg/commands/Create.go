package commands

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/simplecontainer/smr/pkg/bootstrap"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/helpers"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/viper"
	"net"
	"os"
	"strings"
)

func Create() {
	Commands = append(Commands, Command{
		name: "create",
		condition: func(*api.Api) bool {
			return true
		},
		functions: []func(*api.Api, []string){
			func(api *api.Api, args []string) {
				_, err := bootstrap.CreateProject(static.ROOTSMR, api.Config)

				if err != nil {
					panic(err)
				}

				hostHomeDir := ""
				if os.Getenv("HOMEDIR") != "" {
					hostHomeDir = os.Getenv("HOMEDIR")
				}

				hostname := ""
				if os.Getenv("HOSTNAME") != "" {
					hostname = os.Getenv("HOSTNAME")
				}

				api.Config.Platform = viper.GetString("platform")
				api.Config.OverlayNetwork = viper.GetString("overlay")
				api.Config.HostPort.Host, api.Config.HostPort.Port, err = net.SplitHostPort(viper.GetString("port"))

				if err != nil {
					panic(err)
				}

				api.Config.Node = viper.GetString("agent")
				api.Config.Target = viper.GetString("target")
				api.Config.Root = api.Config.Environment.PROJECTDIR
				api.Config.Certificates.Domains = configuration.NewDomains(strings.FieldsFunc(viper.GetString("domains"), helpers.SplitClean))
				api.Config.Certificates.IPs = configuration.NewIPs(strings.FieldsFunc(viper.GetString("ips"), helpers.SplitClean))
				api.Config.HostHome = hostHomeDir
				api.Config.Node = hostname

				// Internal domains needed
				api.Config.Certificates.Domains.Add("localhost")
				api.Config.Certificates.Domains.Add(fmt.Sprintf("smr-agent.%s", static.SMR_LOCAL_DOMAIN))

				// Internal IPs needed
				api.Config.Certificates.IPs.Add("127.0.0.1")

				api.Config.KVStore = &configuration.KVStore{
					Cluster:     []*node.Node{},
					Node:        uint64(viper.GetInt("node")),
					URL:         viper.GetString("url"),
					JoinCluster: viper.GetBool("join"),
				}

				err = startup.Save(api.Config)

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

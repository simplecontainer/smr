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

				api.Config.Platform = viper.GetString("platform")
				api.Config.HostPort.Host, api.Config.HostPort.Port, err = net.SplitHostPort(viper.GetString("port"))

				if err != nil {
					panic(err)
				}

				api.Config.NodeName = viper.GetString("node")
				api.Config.NodeImage = fmt.Sprintf("%s:%s", viper.GetString("image"), viper.GetString("tag"))
				api.Config.Certificates.Domains = configuration.NewDomains(strings.FieldsFunc(viper.GetString("domains"), helpers.SplitClean))
				api.Config.Certificates.IPs = configuration.NewIPs(strings.FieldsFunc(viper.GetString("ips"), helpers.SplitClean))

				// Internal domains needed
				api.Config.Certificates.Domains.Add("localhost")
				api.Config.Certificates.Domains.Add(fmt.Sprintf("%s.%s", static.SMR_ENDPOINT_NAME, static.SMR_LOCAL_DOMAIN))

				// Internal IPs needed
				api.Config.Certificates.IPs.Add("127.0.0.1")

				api.Config.KVStore = &configuration.KVStore{
					Cluster:     []*node.Node{},
					Node:        nil,
					URL:         viper.GetString("url"),
					JoinCluster: viper.GetString("join"),
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

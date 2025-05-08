package commands

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/internal/definitions"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/simplecontainer/smr/pkg/bootstrap"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net"
	"os"
	"strings"
)

func Node() {
	Commands = append(Commands,
		command.Engine{
			Parent: "smr",
			Name:   "node",
			Condition: func(*api.Api) bool {
				return true
			},
			Functions: []func(*api.Api, []string){
				func(api *api.Api, args []string) {

				},
			},
			DependsOn: []func(*api.Api, []string){
				func(api *api.Api, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("flannel.backend", "wireguard", "Flannel backend: vxlan, wireguard")
				cmd.Flags().String("flannel.cidr", "10.10.0.0/16", "Flannel overlay network CIDR")
				cmd.Flags().String("flannel.iface", "", "Network interface for flannel to use, if ommited default gateway will be used")
			},
		},
		command.Engine{
			Parent: "node",
			Name:   "start",
			Condition: func(*api.Api) bool {
				return true
			},
			Functions: []func(*api.Api, []string){
				func(api *api.Api, args []string) {
					environment := configuration.NewEnvironment(configuration.WithHostConfig())
					conf, err := startup.Load(environment)

					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}

					fmt.Println(environment)
					fmt.Println(conf)

					definition, err := definitions.Node(conf.NodeName, conf)

					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}

					var container platforms.IPlatform

					switch conf.Platform {
					case static.PLATFORM_DOCKER:
						if err = docker.IsDaemonRunning(); err != nil {
							fmt.Println(err)
							os.Exit(1)
						}

						container, err = docker.New(conf.NodeName, definition)
						break
					default:
						helpers.PrintAndExit(errors.New("platform not supported"), 1)
					}

					state, err := container.GetState()

					switch state.State {
					case "running":
						helpers.PrintAndExit(errors.New("container is already running"), 1)
						break
					default:
						err = container.Delete()

						if err != nil {
							helpers.PrintAndExit(err, 1)
						}
					}

					err = container.Run()

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					fmt.Println("node started - waiting to be ready...")
				},
			},
			DependsOn: []func(*api.Api, []string){
				func(api *api.Api, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("node", "simplecontainer-node-1", "Node container name")
				viper.BindPFlag("node", cmd.Flags().Lookup("node"))
			},
		},
		command.Engine{
			Parent: "node",
			Name:   "stop",
			Condition: func(*api.Api) bool {
				return true
			},
			Functions: []func(*api.Api, []string){
				func(api *api.Api, args []string) {
					conf, err := startup.Load(configuration.NewEnvironment(configuration.WithHostConfig()))

					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}

					definition, err := definitions.Node(conf.NodeName, conf)

					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}

					var container platforms.IPlatform

					switch api.Config.Platform {
					case static.PLATFORM_DOCKER:
						if err = docker.IsDaemonRunning(); err != nil {
							fmt.Println(err)
							os.Exit(1)
						}

						container, err = docker.New(conf.NodeName, definition)
						break
					default:
						helpers.PrintAndExit(errors.New("container is already running"), 1)
					}

					err = container.Stop(static.SIGTERM)

					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}

					fmt.Println("node started")
				},
			},
			DependsOn: []func(*api.Api, []string){
				func(api *api.Api, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("node", "simplecontainer-node-1", "Node container name")
				viper.BindPFlag("node", cmd.Flags().Lookup("node"))
			},
		},
		command.Engine{
			Parent: "node",
			Name:   "create",
			Condition: func(*api.Api) bool {
				return true
			},
			Functions: []func(*api.Api, []string){
				func(api *api.Api, args []string) {
					// Override home when creating configuration for specific node
					environment := configuration.NewEnvironment(configuration.WithHostConfig())
					_, err := bootstrap.CreateProject(static.ROOTSMR, environment)

					if err != nil {
						panic(err)
					}

					api.Config.Platform = viper.GetString("platform")
					api.Config.HostPort.Host, api.Config.HostPort.Port, err = net.SplitHostPort(viper.GetString("port"))

					if err != nil {
						panic(err)
					}

					api.Config.NodeName = viper.GetString("node")
					api.Config.NodeImage = viper.GetString("image")
					api.Config.NodeTag = viper.GetString("tag")
					api.Config.Certificates.Domains = configuration.NewDomains(strings.FieldsFunc(viper.GetString("domains"), helpers.SplitClean))
					api.Config.Certificates.IPs = configuration.NewIPs(strings.FieldsFunc(viper.GetString("ips"), helpers.SplitClean))

					// Internal domains needed
					api.Config.Certificates.Domains.Add("localhost")
					api.Config.Certificates.Domains.Add(fmt.Sprintf("%s.%s", static.SMR_ENDPOINT_NAME, static.SMR_LOCAL_DOMAIN))

					// Internal IPs needed
					api.Config.Certificates.IPs.Add("127.0.0.1")

					api.Config.KVStore = &configuration.KVStore{
						Cluster: []*node.Node{},
						Node:    nil,
						URL:     viper.GetString("url"),
						Join:    viper.GetBool("join"),
						Peer:    viper.GetString("peer"),
					}

					api.Config.Ports = &configuration.Ports{
						Control: viper.GetString("port.control"),
						Overlay: viper.GetString("port.overlay"),
						Etcd:    viper.GetString("port.etcd"),
					}

					err = startup.Save(api.Config, environment)

					if err != nil {
						panic(err)
					}

					fmt.Println(fmt.Sprintf("config created and saved at %s/config/config.yaml", environment.NodeDirectory))
				},
			},
			DependsOn: []func(*api.Api, []string){
				func(api *api.Api, args []string) {

				},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("platform", static.PLATFORM_DOCKER, "Container platform to manage containers lifecycle")

				cmd.Flags().String("node", "simplecontainer-node-1", "Node container name")

				cmd.Flags().String("image", "quay.io/simplecontainer/smr", "Node image name")
				cmd.Flags().String("tag", "latest", "Node image tag")
				cmd.Flags().String("entrypoint", "/opt/smr/smr", "Entrypoint for the smr")
				cmd.Flags().String("args", "start", "args")
				cmd.Flags().String("raft", "", "Raft Api")
				cmd.Flags().String("peer", "", "Peer for entering cluster first time. Format: https://host:port")
				cmd.Flags().Bool("join", false, "Join the raft")

				cmd.Flags().String("port", "0.0.0.0:1443", "Simplecontainer mTLS listening interface and port combo")
				cmd.Flags().String("domains", "", "Domains that TLS certificates are valid for")
				cmd.Flags().String("ips", "", "IP addresses that TLS certificates are valid for")

				cmd.Flags().String("port.control", ":1443", "Port mapping of node control plane -> Default 0.0.0.0:1443")
				cmd.Flags().String("port.overlay", ":9212", "Port mapping of node overlay raft port  -> Default 0.0.0.0:9212")
				cmd.Flags().String("port.etcd", "2379", "Port mapping of node overlay raft port  -> Default 127.0.0.1:2379 (Cant be exposed to outside!)")

				viper.BindPFlag("platform", cmd.Flags().Lookup("platform"))
				viper.BindPFlag("node", cmd.Flags().Lookup("node"))
				viper.BindPFlag("image", cmd.Flags().Lookup("image"))
				viper.BindPFlag("tag", cmd.Flags().Lookup("tag"))
				viper.BindPFlag("entrypoint", cmd.Flags().Lookup("entrypoint"))
				viper.BindPFlag("args", cmd.Flags().Lookup("args"))
				viper.BindPFlag("raft", cmd.Flags().Lookup("raft"))
				viper.BindPFlag("peer", cmd.Flags().Lookup("peer"))
				viper.BindPFlag("join", cmd.Flags().Lookup("join"))

				viper.BindPFlag("port", cmd.Flags().Lookup("port"))
				viper.BindPFlag("domains", cmd.Flags().Lookup("domains"))
				viper.BindPFlag("ips", cmd.Flags().Lookup("ips"))

				viper.BindPFlag("port.control", cmd.Flags().Lookup("port.control"))
				viper.BindPFlag("port.overlay", cmd.Flags().Lookup("port.overlay"))
				viper.BindPFlag("port.etcd", cmd.Flags().Lookup("port.etcd"))
			},
		},
	)
}

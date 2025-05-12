package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/internal/definitions"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/simplecontainer/smr/pkg/bootstrap"
	"github.com/simplecontainer/smr/pkg/client"
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
	"time"
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
			},
		},
		command.Engine{
			Parent: "node",
			Name:   "start",
			Condition: func(*api.Api) bool {
				return true
			},
			Args: cobra.NoArgs,
			Functions: []func(*api.Api, []string){
				func(api *api.Api, args []string) {
					environment := configuration.NewEnvironment(configuration.WithHostConfig())
					conf, err := startup.Load(environment)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					definition, err := definitions.Node(conf.NodeName, conf)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					var container platforms.IPlatform

					switch conf.Platform {
					case static.PLATFORM_DOCKER:
						if err = docker.IsDaemonRunning(); err != nil {
							helpers.PrintAndExit(err, 1)
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

					fmt.Println(fmt.Sprintf("starting node with the user: %s", definition.Spec.User))
					err = container.Run()

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					go func() {
						ctx, _ := context.WithCancel(context.Background())
						logs, err := container.Logs(ctx, true)

						if err != nil {
							helpers.PrintAndExit(err, 1)
						}

						defer logs.Close()

						scanner := bufio.NewScanner(logs)
						for scanner.Scan() {
							line := scanner.Text()
							fmt.Println(line)
						}

						if err := scanner.Err(); err != nil {
							helpers.PrintAndExit(err, 1)
						}
					}()

					fmt.Println("node started")

					ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
					defer cancel()

					err = helpers.WaitForFileToAppear(ctx, fmt.Sprintf("%s/.ssh/%s.pem", environment.NodeDirectory, conf.NodeName), 500*time.Millisecond)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					var bundle []byte
					bundle, err = os.ReadFile(fmt.Sprintf("%s/.ssh/%s.pem", environment.NodeDirectory, conf.NodeName))

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					_, port, err := net.SplitHostPort(conf.Ports.Control)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					var credentials *client.Credentials
					credentials, err = client.BundleToCredentials(bundle)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					cli := client.New(conf, environment.NodeDirectory)
					cli.Context, err = cli.Manager.CreateContext(conf.NodeName, fmt.Sprintf("https://localhost:%s", port), credentials)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					err = cli.Context.Connect(context.Background(), true)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					if err = cli.Context.Save(); err != nil {
						helpers.PrintAndExit(err, 1)
					}

					err = cli.Manager.SetActive(cli.Context.Name)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					fmt.Println("context saved")
				},
			},
			DependsOn: []func(*api.Api, []string){
				func(api *api.Api, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("node", "simplecontainer-node-1", "Node container name")
				cmd.Flags().String("entrypoint", "/opt/smr/smr", "Entrypoint")
				cmd.Flags().String("args", "start", "Args")
				cmd.Flags().BoolP("y", "y", false, "Say yes to overwrite of context")
			},
		},
		command.Engine{
			Parent: "node",
			Name:   "clean",
			Condition: func(*api.Api) bool {
				return true
			},
			Args: cobra.NoArgs,
			Functions: []func(*api.Api, []string){
				func(api *api.Api, args []string) {
					conf, err := startup.Load(configuration.NewEnvironment(configuration.WithHostConfig()))

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					definition, err := definitions.Node(conf.NodeName, conf)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					var container platforms.IPlatform

					switch conf.Platform {
					case static.PLATFORM_DOCKER:
						if err = docker.IsDaemonRunning(); err != nil {
							helpers.PrintAndExit(err, 1)
						}

						container, err = docker.New(conf.NodeName, definition)
						break
					default:
						helpers.PrintAndExit(errors.New("platform unknown"), 1)
					}

					defer func() {
						err = container.Delete()

						if err != nil {
							helpers.PrintAndExit(err, 1)
						}

						fmt.Println("node deleted")
					}()

					err = container.Stop(static.SIGTERM)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					fmt.Println("node stopped")
				},
			},
			DependsOn: []func(*api.Api, []string){
				func(api *api.Api, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("node", "simplecontainer-node-1", "Node container name")
			},
		},
		command.Engine{
			Parent: "node",
			Name:   "create",
			Condition: func(*api.Api) bool {
				return true
			},
			Args: cobra.NoArgs,
			Functions: []func(*api.Api, []string){
				func(api *api.Api, args []string) {
					environment := configuration.NewEnvironment(configuration.WithHostConfig())

					_, err := bootstrap.CreateProject(static.ROOTSMR, environment, 0750)

					if err != nil {
						panic(err)
					}

					api.Config.Platform = viper.GetString("platform")
					api.Config.HostPort.Host, api.Config.HostPort.Port, err = net.SplitHostPort(viper.GetString("listen"))

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

					err = startup.Save(api.Config, environment, 0750)

					if err != nil {
						panic(err)
					}

					fmt.Println(fmt.Sprintf("config created and saved at %s/config/config.yaml", environment.NodeDirectory))
				},
			},
			DependsOn: []func(*api.Api, []string){
				func(api *api.Api, args []string) {},
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

				cmd.Flags().String("listen", "0.0.0.0:1443", "Simplecontainer mTLS listening interface and port combo")
				cmd.Flags().String("domains", "", "Domains that TLS certificates are valid for")
				cmd.Flags().String("ips", "", "IP addresses that TLS certificates are valid for")

				cmd.Flags().String("port.control", ":1443", "Port mapping of node control plane -> Default 0.0.0.0:1443")
				cmd.Flags().String("port.overlay", ":9212", "Port mapping of node overlay raft port  -> Default 0.0.0.0:9212")
				cmd.Flags().String("port.etcd", "2379", "Port mapping of node overlay raft port  -> Default 127.0.0.1:2379 (Cant be exposed to outside!)")
			},
		},
		command.Engine{
			Parent: "node",
			Name:   "networks",
			Condition: func(*api.Api) bool {
				return true
			},
			Functions: []func(*api.Api, []string){
				func(api *api.Api, args []string) {
					conf, err := startup.Load(configuration.NewEnvironment(configuration.WithHostConfig()))

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					definition, err := definitions.Node(conf.NodeName, conf)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					var container platforms.IPlatform

					switch conf.Platform {
					case static.PLATFORM_DOCKER:
						if err = docker.IsDaemonRunning(); err != nil {
							helpers.PrintAndExit(err, 1)
						}

						container, err = docker.New(conf.NodeName, definition)
						break
					default:
						helpers.PrintAndExit(errors.New("platform unknown"), 1)
					}

					_, err = container.GetState()
					err = container.SyncNetwork()

					networks := container.GetNetwork()

					val, ok := networks[viper.GetString("network")]

					if ok {
						fmt.Println(val)
					}
				},
			},
			DependsOn: []func(*api.Api, []string){
				func(api *api.Api, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("node", "simplecontainer-node-1", "Node")
				cmd.Flags().String("network", "bridge", "Network name")
			},
		},
	)
}

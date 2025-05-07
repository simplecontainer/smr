package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/control"
	"github.com/simplecontainer/smr/pkg/control/controls/start"
	"github.com/simplecontainer/smr/pkg/flannel"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

func Agent() {
	Commands = append(Commands,
		command.Engine{
			Parent: "smr",
			Name:   "agent",
			Condition: func(*api.Api) bool {
				return true
			},
			Functions: []func(*api.Api, []string){
				func(api *api.Api, args []string) {},
			},
			DependsOn: []func(*api.Api, []string){
				func(api *api.Api, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("log", "info", "Log level")
			},
		},
		command.Engine{
			Parent: "agent",
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

					var bundle []byte
					bundle, err = os.ReadFile(fmt.Sprintf("%s/.ssh/%s.pem", environment.NodeDirectory, conf.NodeName))

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					_, port, err := net.SplitHostPort(conf.Ports.Control)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					cli := client.New(conf)

					cli.Context.ApiURL = fmt.Sprintf("https://localhost:%s", port)
					err = cli.Context.GenerateHttpClient(bundle)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					batch := control.NewCommandBatch()

					// Flag raft holds api of the raft
					batch.AddCommand(start.NewStartCommand(map[string]string{
						"raft":    viper.GetString("raft"),
						"cidr":    conf.Flannel.CIDR,
						"backend": conf.Flannel.Backend,
					}))

					var data []byte
					data, err = json.Marshal(batch)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					response := network.Send(cli.Context.Client, fmt.Sprintf("%s/api/v1/cluster/start", fmt.Sprintf("https://localhost:%s", port)), http.MethodPost, data)

					if response.HttpStatus == http.StatusOK || response.ErrorExplanation == static.CLUSTER_STARTED {
						if response.HttpStatus == http.StatusOK {
							fmt.Println(response.Explanation)
						} else {
							fmt.Println(response.ErrorExplanation)
							fmt.Println("trying to run flannel if not running")
						}

						err = helpers.AcquireLock("/var/run/flannel/flannel.lock")

						if err != nil {
							helpers.PrintAndExit(err, 1)
						}

						defer func() {
							err = helpers.ReleaseLock("/var/run/flannel/flannel.lock")
							if err != nil {
								logger.Log.Error("failed to clear lock /var/run/flannel/flannel.lock - do it manually", zap.Error(err))
							}
						}()

						err = os.WriteFile("/var/run/flannel.pid", []byte(fmt.Sprintf("%d", os.Getpid())), 0644)

						if err != nil {
							helpers.PrintAndExit(err, 1)
						}

						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						done := make(chan error, 1)

						go func() {
							logger.Log.Info("starting flannel")
							err = flannel.Run(ctx, cancel, cli, conf)

							if err != nil {
								logger.Log.Error("flannel error:", zap.Error(err))
							} else {
								logger.Log.Info("flannel exited - bye bye")
							}

							done <- err
						}()

						select {
						case <-ctx.Done():
							logger.Log.Info("agent exited: context canceled")
						case err = <-done:
							if err != nil {
								logger.Log.Error("agent exited: flannel exited with error", zap.Error(err))
							} else {
								logger.Log.Info("agent exited: flannel exited with nil")
							}
						}
					} else {
						fmt.Println(response.ErrorExplanation)
					}
				},
			},
			DependsOn: []func(*api.Api, []string){
				func(api *api.Api, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("raft", "", "raft endpoint")

				viper.BindPFlag("raft", cmd.Flags().Lookup("raft"))
			},
		},
		command.Engine{
			Parent: "agent",
			Name:   "stop",
			Condition: func(*api.Api) bool {
				return true
			},
			Functions: []func(*api.Api, []string){
				func(api *api.Api, args []string) {
					pidStr, err := os.ReadFile("/var/run/flannel.pid")

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					var pid int
					pid, err = strconv.Atoi(string(pidStr))

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					var proc *os.Process
					proc, err = os.FindProcess(pid)

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}

					err = proc.Kill()

					if err != nil {
						helpers.PrintAndExit(err, 1)
					} else {
						fmt.Println("process killed successfully")
					}
				},
			},
			DependsOn: []func(*api.Api, []string){
				func(api *api.Api, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("log", "info", "Log level")
			},
		},
		command.Engine{
			Parent: "agent",
			Name:   "control",
			Condition: func(*api.Api) bool {
				return true
			},
			Functions: []func(*api.Api, []string){
				func(api *api.Api, args []string) {
					cli, err := clientv3.New(clientv3.Config{
						Endpoints:   []string{fmt.Sprintf("localhost:%s", api.Config.Ports.Etcd)},
						DialTimeout: 5 * time.Second,
					})

					if err != nil {
						helpers.PrintAndExit(err, 1)
					}
					defer cli.Close()

					logger.Log.Info("listening for control events...")
					watchCh := cli.Watch(context.Background(), "/smr/control/", clientv3.WithPrefix())

					c := client.New(api.Config)

					for watchResp := range watchCh {
						for _, event := range watchResp.Events {
							if event.Type != mvccpb.PUT {
								continue
							}

							logger.Log.Info("new control event received")

							batch := control.NewCommandBatch()

							err = json.Unmarshal(event.Kv.Value, batch)

							if err != nil {
								logger.Log.Error("failed to unmarshal control", zap.Error(err))
								continue
							}

							for _, cmd := range batch.GetCommands() {
								err = cmd.Agent(c, cmd.Data())

								if err != nil {
									logger.Log.Error("error executing control", zap.Error(err))
								}
							}
						}
					}
				},
			},
			DependsOn: []func(*api.Api, []string){
				func(api *api.Api, args []string) {},
			},
			Flags: func(cmd *cobra.Command) {
				cmd.Flags().String("log", "info", "Log level")
			},
		},
	)
}

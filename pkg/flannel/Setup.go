package flannel

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/internal/definitions"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/static"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	emptyIPv6Network = "::/0"

	ipv4 = iota
	ipv6
)

func Run(ctx context.Context, cancel context.CancelFunc, c *client.Client, config *configuration.Configuration) error {
	logger.Log.Info("starting flannel with backend", zap.String("backend", config.Flannel.Backend))

	f := New(subnetFile)
	err := f.Clear()

	if err != nil {
		return err
	}

	err = f.SetBackend(config.Flannel.Backend)

	if err != nil {
		return err
	}

	err = f.EnableIPv4(config.Flannel.EnableIPv4)

	if err != nil {
		return err
	}

	err = f.EnableIPv6(config.Flannel.EnableIPv6)

	if err != nil {
		return err
	}

	f.MaskIPv6(config.Flannel.IPv6Masq)

	err = f.SetCIDR(config.Flannel.CIDR)

	if err != nil {
		return err
	}

	err = f.SetInterface(config.Flannel.InterfaceSpecified)

	if err != nil {
		return err
	}

	netMode, err := findNetMode(f.CIDR)
	if err != nil {
		return errors.Wrap(err, "failed to check netMode for flannel")
	}

	go func() {
		err = flannel(ctx, f, f.InterfaceSpecified, f.IPv6Masq, f.NetMode)
		if err != nil {
			logger.Log.Error("flannel exited: %v", zap.Error(err))
		}

		cancel()
	}()

	var cli *clientv3.Client
	cli, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{fmt.Sprintf("localhost:%s", config.Ports.Etcd)},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return err
	}

	watcher := cli.Watch(ctx, "/coreos.com/network/subnets", clientv3.WithPrefix())
	logger.Log.Info("client will wait for flannel to return subnet range")

	recursion := make(map[string]string)

	for {
		select {
		case watchResp, ok := <-watcher:
			if ok {
				for _, event := range watchResp.Events {
					switch event.Type {
					case mvccpb.PUT:
						if strings.Contains(string(event.Kv.Key), "subnet") {
							if event.Kv.Lease != 0 {
								var subnet = Subnet{}
								err = json.Unmarshal(event.Kv.Value, &subnet)

								if err != nil {
									return err
								}

								logger.Log.Info("got subnet ", zap.Any("subnet", subnet))

								switch netMode {
								case ipv4:
									if f.Interface.ExtAddr.String() == subnet.PublicIP {
										logger.Log.Info("adding it as my own subnet", zap.String("subnet", string(event.Kv.Key)))

										split := strings.Split(string(event.Kv.Key), "/")
										CIDR := strings.Replace(split[len(split)-1], "-", "/", 1)

										NetworkDefinition, _ := definitions.ClusterNetwork(CIDR).ToJSON()

										req, err := common.NewRequest(static.KIND_NETWORK)

										if err != nil {
											fmt.Println(err)
											break
										}

										err = req.Definition.FromJson(NetworkDefinition)

										if err != nil {
											fmt.Println(err)
											break
										}

										err = req.ProposeApply(c.Context.Client, c.Context.ApiURL)

										if err != nil {
											fmt.Println(err)
										} else {
											fmt.Println("network object applied")
										}
									}
									break
								case ipv6:
									if f.Interface.ExtV6Addr.String() == subnet.PublicIPv6 {
										split := strings.Split(string(event.Kv.Key), "/")
										CIDR := strings.Replace(split[len(split)-1], "-", "/", 1)

										NetworkDefinition, _ := definitions.ClusterNetwork(CIDR).ToJSON()

										req, err := common.NewRequest(static.KIND_NETWORK)

										if err != nil {
											fmt.Println(err)
											break
										}

										err = req.Definition.FromJson(NetworkDefinition)

										if err != nil {
											fmt.Println(err)
											break
										}

										err = req.ProposeApply(c.Context.Client, c.Context.ApiURL)

										if err != nil {
											fmt.Println(err)
										} else {
											fmt.Println("network object applied")
										}
									}
									break
								case ipv4 | ipv6:
									break
								}

								if recursion[string(event.Kv.Key)] == string(event.Kv.Value) {
									continue
								}

								recursion[string(event.Kv.Key)] = string(event.Kv.Value)
								response := network.Send(c.Context.Client, fmt.Sprintf("%s/api/v1/key/propose/%s", c.Context.ApiURL, event.Kv.Key), http.MethodPost, event.Kv.Value)

								if response.Success {
									go func() {
										var kach <-chan *clientv3.LeaseKeepAliveResponse
										kach, err = cli.KeepAlive(ctx, clientv3.LeaseID(event.Kv.Lease))

										for {
											select {
											case data, ok := <-kach:
												if ok {
													logger.Log.Info(fmt.Sprintf("keep alived: %s", data.String()))
													break
												} else {
													logger.Log.Info(fmt.Sprintf("closed keep alive channel for lease: %s", event.Kv.Lease))
													return
												}
											}
										}
									}()
								} else {
									logger.Log.Error("flannel failed to inform members about subnet decision - abort startup")
									os.Exit(1)
								}
							}
						}

					}
				}
			}
		case <-ctx.Done():
			return errors.New("closed watcher channel - should not block")
		}
	}
}

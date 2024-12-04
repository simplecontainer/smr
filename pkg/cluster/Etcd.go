package cluster

import (
	"context"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	"net/url"
	"time"
)

func NewEtcdClient() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return nil
	}

	return cli
}

func (c *Cluster) StartSingleNodeEtcd(config *configuration.Configuration) (e *embed.Etcd, err error) {
	cfg := embed.NewConfig()
	cfg.Dir = fmt.Sprintf("%s/persistent/etcd", config.Environment.PROJECTDIR)

	URLC, _ := url.Parse("http://0.0.0.0:2379")

	cfg.AdvertiseClientUrls = []url.URL{*URLC}
	cfg.ListenClientUrls = []url.URL{*URLC}
	cfg.Logger = "zap"
	cfg.LogOutputs = []string{"/tmp/etcd.log"}

	return embed.StartEtcd(cfg)
}
func (c *Cluster) ConfigureFlannel(network string) error {
	timeout, err := time.ParseDuration("20s")

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	_, err = c.EtcdClient.Put(ctx, "/coreos.com/network/config", fmt.Sprintf("{\"Network\": \"%s\", \"Backend\": {\"Type\": \"vxlan\"}}", network))
	cancel()

	if err != nil {
		return err
	}

	return nil
}
func (c *Cluster) ListenEvents(agent string) {
	ctx, _ := context.WithCancel(context.Background())
	watcher := c.EtcdClient.Watch(ctx, "/coreos.com", clientv3.WithPrefix())

	for {
		select {
		case watchResp, ok := <-watcher:
			if ok {
				for _, event := range watchResp.Events {
					switch event.Type {
					case mvccpb.PUT:
						fmt.Println(fmt.Sprintf("proposing put changes: %s", agent))
						c.KVStore.ProposeEtcd(string(event.Kv.Key), string(event.Kv.Value), agent)
						break
					case mvccpb.DELETE:
						fmt.Println(fmt.Sprintf("proposing delete changes: %s", agent))
						c.KVStore.ProposeEtcd(string(event.Kv.Key), "", agent)
						break
					}
				}
			}
		case <-ctx.Done():
			logger.Log.Error(errors.New("closed watcher channel should not block").Error())
		}
	}
}

func (c *Cluster) ListenUpdates(agent string) {
	for {
		select {
		case data, ok := <-c.KVStore.EtcdC:
			if ok {
				if data.Agent != agent {
					ctx, _ := context.WithCancel(context.Background())

					val, err := c.EtcdClient.Get(ctx, data.Key)

					if err != nil {
						logger.Log.Error(err.Error())
					}

					if len(val.Kvs) == 0 || string(val.Kvs[len(val.Kvs)-1].Value) != data.Val {
						_, err = c.EtcdClient.Put(ctx, data.Key, data.Val)

						fmt.Println(fmt.Sprintf("etcd insert %s=%s", data.Key, data.Val))

						if err != nil {
							logger.Log.Error(err.Error())
						}
					}
				}
			}
		}
	}
}

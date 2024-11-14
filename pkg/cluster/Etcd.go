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
func (c *Cluster) ListenEvents() {
	ctx, _ := context.WithCancel(context.Background())
	watcher := c.EtcdClient.Watch(ctx, "/coreos.com", clientv3.WithPrefix())

	for {
		select {
		case watchResp, ok := <-watcher:
			if ok {
				for _, event := range watchResp.Events {
					switch event.Type {
					case mvccpb.PUT:
						fmt.Println("proposing put changes")
						c.KVStore.Propose(string(event.Kv.Key), string(event.Kv.Value))
					case mvccpb.DELETE:
						fmt.Println("proposing delete changes")
						c.KVStore.Propose(string(event.Kv.Key), "")
					}
				}
			}
		case <-ctx.Done():
			logger.Log.Error(errors.New("closed watcher channel should not block").Error())
		}
	}
}

package networking

import (
	"context"
	"fmt"
	"github.com/simplecontainer/smr/pkg/configuration"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	"net/url"
	"time"
)

func StartEtcd(config *configuration.Configuration) (e *embed.Etcd, err error) {
	cfg := embed.NewConfig()
	cfg.Dir = fmt.Sprintf("%s/persistent/etcd", config.Environment.PROJECTDIR)

	URLC, _ := url.Parse("http://0.0.0.0:2379")

	cfg.AdvertiseClientUrls = []url.URL{*URLC}
	cfg.ListenClientUrls = []url.URL{*URLC}
	cfg.Logger = "zap"
	cfg.LogOutputs = []string{"/tmp/etcd.log"}

	return embed.StartEtcd(cfg)
}

func Flannel(network string, backend string) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return err
	}

	timeout, err := time.ParseDuration("10s")

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	_, err = cli.Put(ctx, "/coreos.com/network/config", fmt.Sprintf("{\"Network\": \"%s\", \"Backend\": {\"Type\": \"%s\"}}", network, backend))
	cancel()

	if err != nil {
		return err
	}

	return nil
}

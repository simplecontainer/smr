package etcd

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/configuration"
	"go.etcd.io/etcd/server/v3/embed"
	"net/url"
)

func StartEtcd(config *configuration.Configuration) (e *embed.Etcd, err error) {
	cfg := embed.NewConfig()
	cfg.Dir = fmt.Sprintf("%s/persistent/etcd", config.Environment.NodeDirectory)

	URLC, _ := url.Parse("http://0.0.0.0:2379")

	cfg.AdvertiseClientUrls = []url.URL{*URLC}
	cfg.ListenClientUrls = []url.URL{*URLC}

	cfg.SnapshotCount = config.Etcd.SnapshotCount
	cfg.MaxSnapFiles = config.Etcd.MaxSnapFiles
	cfg.MaxWalFiles = config.Etcd.MaxWalFiles

	cfg.AutoCompactionMode = config.Etcd.AutoCompactionMode
	cfg.AutoCompactionRetention = config.Etcd.AutoCompactionRetention

	cfg.QuotaBackendBytes = config.Etcd.QuotaBackendBytes
	cfg.MaxTxnOps = config.Etcd.MaxTxnOps

	cfg.EnableV2 = config.Etcd.EnableV2
	cfg.EnableGRPCGateway = config.Etcd.EnableGRPCGateway

	cfg.Logger = "zap"
	cfg.LogOutputs = []string{fmt.Sprintf("/tmp/etcd-%s.log", config.NodeName)}

	return embed.StartEtcd(cfg)
}

package networking

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
	cfg.SnapshotCount = 1000
	cfg.MaxSnapFiles = 10
	cfg.MaxWalFiles = 10

	cfg.AutoCompactionMode = "revision"
	cfg.AutoCompactionRetention = "1m"

	cfg.QuotaBackendBytes = 8 * 1024 * 1024
	cfg.MaxTxnOps = 64
	cfg.EnableV2 = false
	cfg.EnableGRPCGateway = false
	cfg.Logger = "zap"
	cfg.LogOutputs = []string{fmt.Sprintf("/tmp/etcd-%s.log", config.NodeName)}

	return embed.StartEtcd(cfg)
}

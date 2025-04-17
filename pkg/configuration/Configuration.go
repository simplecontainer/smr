package configuration

import (
	"fmt"
	ips "github.com/simplecontainer/smr/pkg/network/ip"
	"github.com/simplecontainer/smr/pkg/static"
)

func NewConfig() *Configuration {
	IPs, err := ips.NewfromEtcHosts()

	if err != nil {
		panic(err)
	}

	return &Configuration{
		Environment: &Environment{
			Home:          "/home/node",
			NodeDirectory: fmt.Sprintf("%s/%s", "/home/node", static.ROOTDIR),
			NodeIP:        IPs.IPs[len(IPs.IPs)-1].String(),
		},
		Certificates: &Certificates{},
		Etcd:         DefaultEtcdConfig(),
	}
}

func DefaultEtcdConfig() *EtcdConfiguration {
	return &EtcdConfiguration{
		DataDir:                 "/var/lib/etcd",
		QuotaBackendBytes:       8 * 1024 * 1024 * 1024, // 8GB storage
		SnapshotCount:           1000,
		MaxSnapFiles:            10,
		MaxWalFiles:             10,
		AutoCompactionMode:      "revision",
		AutoCompactionRetention: "1",
		MaxTxnOps:               64,
		EnableV2:                false,
		EnableGRPCGateway:       true,
		LoggerType:              "zap",
		LogOutputs:              []string{"/tmp/etcd.log"},
	}
}

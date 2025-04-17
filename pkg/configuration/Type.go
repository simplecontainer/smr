package configuration

import (
	"github.com/simplecontainer/smr/pkg/node"
)

type Configuration struct {
	Platform     string             `yaml:"platform"`
	NodeImage    string             `yaml:"nodeImage"`
	NodeName     string             `yaml:"nodeName"`
	HostPort     HostPort           `yaml:"hostport"`
	KVStore      *KVStore           `yaml:"kvstore"`
	Certificates *Certificates      `yaml:"certificates"`
	Environment  *Environment       `yaml:"-"`
	Etcd         *EtcdConfiguration `yaml:"etcd"`
}

type HostPort struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Environment struct {
	Home          string
	NodeIP        string
	NodeDirectory string
}

type KVStore struct {
	Cluster []*node.Node `yaml:"cluster"`
	Node    *node.Node   `yaml:"node"`
	URL     string       `yaml:"url"`
	API     string       `yaml:"api"`
	Join    bool         `yaml:"join"`
	Peer    string       `yaml:"peer"`
}

type Certificates struct {
	Domains *Domains `yaml:"domains"`
	IPs     *IPs     `yaml:"ips"`
}

type IPs struct {
	Members []string `yaml:"members"`
}

type Domains struct {
	Members []string `yaml:"members"`
}

type EtcdConfiguration struct {
	DataDir           string
	QuotaBackendBytes int64

	SnapshotCount uint64
	MaxSnapFiles  uint
	MaxWalFiles   uint

	AutoCompactionMode      string
	AutoCompactionRetention string

	MaxTxnOps uint

	EnableV2          bool
	EnableGRPCGateway bool

	LoggerType string
	LogOutputs []string
}

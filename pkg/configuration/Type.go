package configuration

import (
	"github.com/simplecontainer/smr/pkg/node"
)

type Configuration struct {
	Environment  *EnvironmentDual      `yaml:"-"`
	Home         string                `yaml:"home"`
	Platform     string                `yaml:"platform"`
	NodeImage    string                `yaml:"nodeImage"`
	NodeTag      string                `yaml:"nodeTag"`
	NodeName     string                `yaml:"nodeName"`
	HostPort     HostPort              `yaml:"hostport"`
	KVStore      *KVStore              `yaml:"kvstore"`
	Certificates *Certificates         `yaml:"certificates"`
	Ports        *Ports                `yaml:"ports"`
	Etcd         *EtcdConfiguration    `yaml:"etcd"`
	Flannel      *FlannelConfiguration `yaml:"flannel"`
}

type HostPort struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type EnvironmentDual struct {
	Container *Environment
	Host      *Environment
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

type Ports struct {
	Control string
	Overlay string
	Etcd    string
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

type FlannelConfiguration struct {
	Backend            string
	CIDR               string
	InterfaceSpecified string
	EnableIPv4         bool
	EnableIPv6         bool
	IPv6Masq           bool
}

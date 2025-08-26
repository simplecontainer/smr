package configuration

import (
	"github.com/simplecontainer/smr/pkg/node"
	"time"
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
	User            string
	Home            string
	NodeIP          string
	NodeDirectory   string
	ClientDirectory string
}

type KVStore struct {
	Cluster []*node.Node `yaml:"cluster"`
	Node    *node.Node   `yaml:"node"`
	URL     string       `yaml:"url"`
	API     string       `yaml:"api"`
	Join    bool         `yaml:"join"`
	Peer    string       `yaml:"peer"`
	Replay  bool         `yaml:"replay"`
}

type Ports struct {
	Control string
	Overlay string
	Etcd    string
	Traefik string
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

var Timeout = NewTimeouts()

func NewTimeouts() *Timeouts {
	return &Timeouts{
		AcknowledgmentTimeout:     10 * time.Second,
		ResourceDrainTimeout:      1800 * time.Second,
		CompleteDrainTimeout:      360 * time.Second,
		EtcdConnectionTimeout:     5 * time.Second,
		NodeStartupTimeout:        60 * time.Second,
		LeadershipTransferTimeout: 60 * time.Second,
	}
}

type Timeouts struct {
	AcknowledgmentTimeout     time.Duration `yaml:"acknowledgment_timeout"`
	ResourceDrainTimeout      time.Duration `yaml:"resource_drain_timeout"`
	CompleteDrainTimeout      time.Duration `yaml:"kind_drain_timeout"`
	EtcdConnectionTimeout     time.Duration `yaml:"etcd_connection_timeout"`
	NodeStartupTimeout        time.Duration `yaml:"node_startup_timeout"`
	LeadershipTransferTimeout time.Duration `yaml:"leadership_transfer_timeout"`
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

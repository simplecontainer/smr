package configuration

import (
	"github.com/simplecontainer/smr/pkg/node"
	"time"
)

type Configuration struct {
	Environment  *EnvironmentDual      `mapstructure:"-"`
	Home         string                `mapstructure:"home"`
	Platform     string                `mapstructure:"platform"`
	NodeImage    string                `mapstructure:"nodeImage"`
	NodeTag      string                `mapstructure:"nodeTag"`
	NodeName     string                `mapstructure:"nodeName"`
	HostPort     HostPort              `mapstructure:"hostport"`
	KVStore      *KVStore              `mapstructure:"kvstore"`
	Certificates *Certificates         `mapstructure:"certificates"`
	Ports        *Ports                `mapstructure:"ports"`
	Etcd         *EtcdConfiguration    `mapstructure:"etcd"`
	RaftConfig   *RaftConfiguration    `mapstructure:"raftConfig"`
	Flannel      *FlannelConfiguration `mapstructure:"flannel"`
}

type HostPort struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type EnvironmentDual struct {
	Container *Environment
	Host      *Environment
}

type Environment struct {
	User            string
	Groups          []string
	Home            string
	NodeIP          string
	NodeDirectory   string
	ClientDirectory string
}

type KVStore struct {
	Cluster []*node.Node `mapstructure:"cluster"`
	Node    *node.Node   `mapstructure:"node"`
	URL     string       `mapstructure:"url"`
	API     string       `mapstructure:"api"`
	Join    bool         `mapstructure:"join"`
	Peer    string       `mapstructure:"peer"`
	Replay  bool         `mapstructure:"replay"`
}

type Ports struct {
	Control string `mapstructure:"control"`
	Overlay string `mapstructure:"overlay"`
	Etcd    string `mapstructure:"etcd"`
	Traefik string `mapstructure:"traefik"`
}

type Certificates struct {
	Domains *Domains `mapstructure:"domains"`
	IPs     *IPs     `mapstructure:"ips"`
}

type IPs struct {
	Members []string `mapstructure:"members"`
}

type Domains struct {
	Members []string `mapstructure:"members"`
}

var Timeout = NewTimeouts()

func NewTimeouts() *Timeouts {
	return &Timeouts{
		AcknowledgmentTimeout:     60 * time.Second,
		ResourceDrainTimeout:      1800 * time.Second,
		CompleteDrainTimeout:      360 * time.Second,
		EtcdConnectionTimeout:     5 * time.Second,
		NodeStartupTimeout:        60 * time.Second,
		LeadershipTransferTimeout: 60 * time.Second,
	}
}

type Timeouts struct {
	AcknowledgmentTimeout     time.Duration `mapstructure:"acknowledgment_timeout"`
	ResourceDrainTimeout      time.Duration `mapstructure:"resource_drain_timeout"`
	CompleteDrainTimeout      time.Duration `mapstructure:"kind_drain_timeout"`
	EtcdConnectionTimeout     time.Duration `mapstructure:"etcd_connection_timeout"`
	NodeStartupTimeout        time.Duration `mapstructure:"node_startup_timeout"`
	LeadershipTransferTimeout time.Duration `mapstructure:"leadership_transfer_timeout"`
}

type EtcdConfiguration struct {
	DataDir                 string   `mapstructure:"datadir"`
	QuotaBackendBytes       int64    `mapstructure:"quotabackendbytes"`
	SnapshotCount           uint64   `mapstructure:"snapshotcount"`
	MaxSnapFiles            uint     `mapstructure:"maxsnapfiles"`
	MaxWalFiles             uint     `mapstructure:"maxwalfiles"`
	AutoCompactionMode      string   `mapstructure:"autocompactionmode"`
	AutoCompactionRetention string   `mapstructure:"autocompactionretention"`
	MaxTxnOps               uint     `mapstructure:"maxtxnops"`
	EnableV2                bool     `mapstructure:"enablev2"`
	EnableGRPCGateway       bool     `mapstructure:"enablegrpcgateway"`
	LoggerType              string   `mapstructure:"loggertype"`
	LogOutputs              []string `mapstructure:"logoutputs"`
}

type RaftConfiguration struct {
	SnapshotCount          uint64        `mapstructure:"snapshotcount"`
	SnapshotCatchUpEntries uint64        `mapstructure:"snapshotcatchupentries"`
	SnapshotInterval       time.Duration `mapstructure:"snapshotinterval"`
	EnablePeriodicCleanup  bool          `mapstructure:"enableperiodiccleanup"`
	CleanupInterval        time.Duration `mapstructure:"cleanupinterval"`
	KeepSnapshotCount      int           `mapstructure:"keepsnapshotcount"`
	EnableWALCleanup       bool          `mapstructure:"enablewalcleanup"`
	ElectionTick           int           `mapstructure:"electiontick"`
	HeartbeatTick          int           `mapstructure:"heartbeattick"`
	MaxSizePerMsg          uint64        `mapstructure:"maxsizepermsg"`
	MaxInflightMsgs        int           `mapstructure:"maxinflightmsgs"`
	MaxUncommittedEntries  uint64        `mapstructure:"maxuncommittedentries"`
	DialTimeout            time.Duration `mapstructure:"dialtimeout"`
	DialRetryFrequency     time.Duration `mapstructure:"dialretryfrequency"`
}

type FlannelConfiguration struct {
	Backend            string `mapstructure:"backend"`
	CIDR               string `mapstructure:"cidr"`
	InterfaceSpecified string `mapstructure:"interfacespecified"`
	EnableIPv4         bool   `mapstructure:"enableipv4"`
	EnableIPv6         bool   `mapstructure:"enableipv6"`
	IPv6Masq           bool   `mapstructure:"ipv6masq"`
}

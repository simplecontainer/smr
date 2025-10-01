package configuration

import (
	"fmt"
	ips "github.com/simplecontainer/smr/pkg/network/ip"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"time"
)

func NewConfig() *Configuration {
	// Viper flags will only be read at the create command
	// The others commands will overrun this and will read from the config file produced by the create command!
	// Flow: create -> save config to yaml file -> start -> read from yaml config file -> run the engine

	return &Configuration{
		Home: viper.GetString("home"),
		Environment: &EnvironmentDual{
			Container: NewEnvironment(WithContainerConfig()),
			Host:      NewEnvironment(WithHostConfig()),
		},
		Certificates: &Certificates{},
		Etcd:         DefaultEtcdConfig(),
		RaftConfig:   DefaultRaftConfig(),
		Flannel:      DefaultFlannelConfig(),
	}
}

func NewEnvironment(opts ...EnvOption) *Environment {
	env := &Environment{}

	for _, opt := range opts {
		opt(env)
	}

	return env
}

type EnvOption func(*Environment)

func WithContainerConfig() EnvOption {
	return func(env *Environment) {
		IPs, err := ips.NewfromEtcHosts()
		if err != nil {
			panic(err)
		}

		env.Home = viper.GetString("home")
		env.NodeDirectory = viper.GetString("home")
		env.NodeIP = IPs.IPs[len(IPs.IPs)-1].String()
	}
}

func WithHostConfig() EnvOption {
	return func(env *Environment) {
		env.User = fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())
		gids, err := os.Getgroups()
		if err != nil {
			panic("failed to get groups")
		}

		for _, gid := range gids {
			env.Groups = append(env.Groups, strconv.Itoa(gid))
		}

		env.Home = viper.GetString("home")
		env.NodeDirectory = fmt.Sprintf("%s/nodes/%s", viper.GetString("home"), viper.GetString("node"))
		env.ClientDirectory = fmt.Sprintf("%s/.smrctl", viper.GetString("home"))
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

func DefaultRaftConfig() *RaftConfiguration {
	return &RaftConfiguration{
		SnapshotCount:          1000,             // Snapshot every 1k entries
		SnapshotCatchUpEntries: 10,               // Keep only 10 entries after snapshot
		SnapshotInterval:       10 * time.Minute, // Force snapshot every 10 minutes

		EnablePeriodicCleanup: true,
		CleanupInterval:       10 * time.Minute,
		KeepSnapshotCount:     3, // Keep last 3 snapshots
		EnableWALCleanup:      true,

		ElectionTick:  10,
		HeartbeatTick: 1,

		MaxSizePerMsg:         1024 * 1024, // 1MB
		MaxInflightMsgs:       256,
		MaxUncommittedEntries: 1 << 30, // 1GB

		DialTimeout:        5 * time.Second,
		DialRetryFrequency: 300 * time.Millisecond,
	}
}

func DefaultFlannelConfig() *FlannelConfiguration {
	return &FlannelConfiguration{
		Backend:            "wireguard",
		CIDR:               "10.10.0.0/16",
		InterfaceSpecified: "",
		EnableIPv4:         false,
		EnableIPv6:         false,
		IPv6Masq:           false,
	}
}

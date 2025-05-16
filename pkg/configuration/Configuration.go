package configuration

import (
	"fmt"
	ips "github.com/simplecontainer/smr/pkg/network/ip"
	"github.com/spf13/viper"
	"os"
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

package startup

import (
	"flag"
	"fmt"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"os"
)

func Load(environment *configuration.Environment) (*configuration.Configuration, error) {
	path := fmt.Sprintf("%s/%s/config.yaml", environment.NodeDirectory, static.CONFIGDIR)

	file, err := os.Open(path)

	defer func() {
		file.Close()
	}()

	if err != nil {
		return nil, err
	}

	configObj := configuration.NewConfig()

	viper.SetConfigType("yaml")
	err = viper.ReadConfig(file)

	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(configObj)

	if err != nil {
		return nil, err
	}

	return configObj, err
}

func Save(configObj *configuration.Configuration) error {
	yamlObj, err := yaml.Marshal(*configObj)

	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%s/%s/config.yaml", configObj.Environment.Home, static.ROOTSMR, static.CONFIGDIR)

	err = os.WriteFile(path, yamlObj, 0644)
	if err != nil {
		return err
	}

	return nil
}

func SetFlags() {
	flag.String("port", "0.0.0.0:1443", "Simplecontainer TLS listening interface and port")
	flag.String("platform", static.PLATFORM_DOCKER, "Container platform to manage containers lifecycle")
	flag.String("domains", "", "Domains that TLS certificates are valid for")
	flag.String("ips", "", "IP addresses that TLS certificates are valid for")

	flag.String("node", "", "Node container name")
	flag.Int("id", 0, "Distributed KVStore Node ID")

	flag.String("image", "", "Node image name")
	flag.String("tag", "", "Node image tag")
	flag.String("cluster", "", "SMR Cluster")
	flag.Bool("join", false, "Join the cluster")
	flag.Bool("restore", false, "Restore cluster")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
}

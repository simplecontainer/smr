package startup

import (
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

func Save(config *configuration.Configuration, environment *configuration.Environment) error {
	yamlObj, err := yaml.Marshal(*config)

	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%s/config.yaml", environment.NodeDirectory, static.CONFIGDIR)

	err = os.WriteFile(path, yamlObj, 0644)
	if err != nil {
		return err
	}

	return nil
}

func EngineFlags() {
	earlyFlags := pflag.NewFlagSet("early", pflag.ContinueOnError)

	earlyFlags.String("home", "/home/node", "Root directory for all actions - keep default inside container")
	earlyFlags.String("log", "info", "Log level: debug, info, warn, error, dpanic, panic, fatal")
	earlyFlags.Bool("y", false, "Say yes to everything")

	viper.BindPFlag("home", earlyFlags.Lookup("home"))
	viper.BindPFlag("log", earlyFlags.Lookup("log"))
}

func ClientFlags() {
	earlyFlags := pflag.NewFlagSet("early", pflag.ContinueOnError)

	earlyFlags.String("log", "info", "Log level: debug, info, warn, error, dpanic, panic, fatal")
	earlyFlags.String("g", "", "Group")

	viper.BindPFlag("log", earlyFlags.Lookup("log"))
	viper.BindPFlag("g", earlyFlags.Lookup("g"))
}

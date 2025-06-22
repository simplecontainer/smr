package startup

import (
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
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

func Save(config *configuration.Configuration, environment *configuration.Environment, permissions os.FileMode) error {
	yamlObj, err := yaml.Marshal(*config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	path := filepath.Join(environment.NodeDirectory, static.CONFIGDIR, "config.yaml")

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(path, yamlObj, permissions)
}

func EngineFlags() {
	// These are only available in the main before cobra starts parsing flags
	// environment := configuration.NewEnvironment(configuration.WithHostConfig()) will place root dir correctly
	// with information provided by these flags - leave default if not sure: it will use home directory as root
	earlyFlags := pflag.NewFlagSet("early", pflag.ContinueOnError)

	earlyFlags.String("home", helpers.GetRealHome(), "Root directory for all actions - keep default inside container")
	earlyFlags.String("log", "info", "Log level: debug, info, warn, error, dpanic, panic, fatal")

	viper.BindPFlag("home", earlyFlags.Lookup("home"))
	viper.BindPFlag("log", earlyFlags.Lookup("log"))
}

func ClientFlags() {
	// These are only available in the main before cobra starts parsing flags
	// environment := configuration.NewEnvironment(configuration.WithHostConfig()) will place root dir correctly
	// with information provided by these flags - leave default if not sure: it will use home directory as root
	earlyFlags := pflag.NewFlagSet("early", pflag.ContinueOnError)

	earlyFlags.String("home", helpers.GetRealHome(), "Root directory for all actions - keep default inside container")
	earlyFlags.String("log", "info", "Log level: debug, info, warn, error, dpanic, panic, fatal")

	viper.BindPFlag("home", earlyFlags.Lookup("home"))
	viper.BindPFlag("log", earlyFlags.Lookup("log"))
}

package config

import (
	"flag"
	"fmt"
	"github.com/qdnqn/smr/pkg/logger"
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Configuration *Configuration
}

func NewConfig() *Config {
	config := Configuration{}

	return &Config{
		Configuration: &config,
	}
}

func (c *Config) Load(projectDir string) {
	configArg := viper.GetString("config")

	if os.Getenv("CONFIG_ARGUMENT") != "" {
		configArg = os.Getenv("CONFIG_ARGUMENT")
	} else {
		if configArg == "" {
			configArg = viper.GetString("project")
		}
	}

	viper.SetConfigName(configArg)
	viper.AddConfigPath(fmt.Sprintf("%s/%s", projectDir, "config"))
	c.Configuration.Environment.Target = configArg

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	err = viper.Unmarshal(c.Configuration)
	if err != nil {
		panic(fmt.Errorf("fatal unable to unmarshal config file: %w", err))
	}
}

func (c *Config) Save(projectDir string) {
	configArg := viper.GetString("config")

	if os.Getenv("CONFIG_ARGUMENT") != "" {
		configArg = os.Getenv("CONFIG_ARGUMENT")
	} else {
		if configArg == "" {
			configArg = fmt.Sprintf("%s.conf", viper.GetString("project"))
		}
	}

	replica := *c.Configuration

	yaml, err := yaml.Marshal(replica)

	if err != nil {
		logger.Log.Fatal("Error while Marshaling. %v", zap.Error(err))
	}

	d1 := []byte(yaml)
	err = os.WriteFile(fmt.Sprintf("%s/%s/%s.yaml", projectDir, "config", configArg), d1, 0644)

	if err != nil {
		panic(err)
	}
}

func (c *Config) ReadFlags() {
	/* Operation mode */
	flag.Bool("daemon", false, "Run daemon as HTTP API")
	flag.Bool("daemon-secured", false, "Run daemon as HTTPS mTLS API")

	/* Client cli config options */
	flag.String("context", "default", "Context file to use for connection")

	/* Meta data */
	flag.String("project", "", "Project name to operate on")
	flag.Bool("optmode", false, "Project is setup in the /opt/smr directory act accordingly")
	flag.Bool("context", false, "Create context file holding current project information for ease of use")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)
}

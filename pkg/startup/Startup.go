package startup

import (
	"flag"
	"fmt"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"os"
)

func Load(configObj *configuration.Configuration, projectDir string) {
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
	configObj.Target = configArg

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	err = viper.Unmarshal(configObj)
	if err != nil {
		panic(fmt.Errorf("fatal unable to unmarshal config file: %w", err))
	}
}

func Save(configObj *configuration.Configuration, projectDir string) {
	configArg := viper.GetString("config")

	if os.Getenv("CONFIG_ARGUMENT") != "" {
		configArg = os.Getenv("CONFIG_ARGUMENT")
	} else {
		if configArg == "" {
			configArg = fmt.Sprintf("%s.conf", viper.GetString("project"))
		}
	}

	replica := *configObj

	yaml, err := yaml.Marshal(replica)

	if err != nil {
		panic(err)
	}

	d1 := []byte(yaml)
	err = os.WriteFile(fmt.Sprintf("%s/%s/%s.yaml", projectDir, "config", configArg), d1, 0644)

	if err != nil {
		panic(err)
	}
}

func ReadFlags(configObj *configuration.Configuration) {
	/* Operation mode */
	flag.Bool("daemon", false, "Run daemon as HTTP API")
	flag.Bool("daemon-secured", false, "Run daemon as HTTPS mTLS API")
	flag.String("daemon-domain", "localhost", "Domain name where daemon will be exposed to")

	/* Client cli config options */
	flag.String("context", "default", "Context file to use for connection")

	/* Logs and output */
	flag.Bool("verbose", false, "Verbose output of the cli and daemon")

	/* Meta data */
	flag.String("project", "", "Project name to operate on")
	flag.Bool("optmode", false, "Project is setup in the /opt/smr directory act accordingly")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	viper.BindPFlags(pflag.CommandLine)

	configObj.Flags.Daemon = viper.GetBool("daemon")
	configObj.Flags.DaemonSecured = viper.GetBool("daemon-secured")
	configObj.Flags.DaemonDomain = viper.GetString("daemon-domain")
	configObj.Flags.OptMode = viper.GetBool("optmode")
	configObj.Flags.Verbose = viper.GetBool("verbose")
}

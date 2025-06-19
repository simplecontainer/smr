package main

import (
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/engine/commands"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/version"
	"github.com/spf13/viper"
)

func main() {
	startup.EngineFlags()
	logger.Log = logger.NewLogger(viper.GetString("log"), []string{"stdout"}, []string{"stderr"})

	if viper.GetString("log") == "debug" {
		fmt.Println(fmt.Sprintf("logging level set to %s (override with --log flag)", viper.GetString("log")))
	}

	// Create configuration for the commands
	conf := configuration.NewConfig()

	// Init the api with proper configuration
	a := api.NewApi(conf)
	a.Version = version.New("", SMR_VERSION)
	a.Manager.LogLevel = helpers.GetLogLevel(viper.GetString("log"))

	cmd := command.New()
	commands.PreloadCommands()
	commands.Parse(a, cmd)
	commands.Run(cmd)
}

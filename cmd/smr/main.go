package main

import (
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/engine/commands"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/simplecontainer/smr/pkg/version"
	"github.com/spf13/cobra"
	_ "net/http/pprof"
	"os"
)

func main() {
	startup.EngineFlags()

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = static.DEFAULT_LOG_LEVEL
	}

	logger.Log = logger.NewLogger(logLevel, []string{"stdout"}, []string{"stderr"})

	if logLevel == "debug" {
		fmt.Println(fmt.Sprintf("logging level set to %s (override with LOG_LEVEL env variable or --log flag)", logLevel))
	}

	// Create configuration for the commands
	conf := configuration.NewConfig()

	// Init the api with proper configuration
	api := api.NewApi(conf)
	api.Version = version.New("", SMR_VERSION)
	api.Manager.LogLevel = helpers.GetLogLevel(logLevel)

	cmd := &cobra.Command{
		Use:   "smr",
		Short: "SMR CLI",
	}

	commands.PreloadCommands()
	commands.Run(api, cmd)
}

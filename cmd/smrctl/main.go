package main

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/client/commands"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/simplecontainer/smr/pkg/version"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	startup.ClientFlags()

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

	// Init the client
	c := client.New(conf)
	c.Version = version.NewClient(SMR_VERSION)

	cmd := &cobra.Command{
		Use:   "smrctl",
		Short: "SMR CLI",
	}

	commands.PreloadCommands()
	commands.Run(c, cmd)
}

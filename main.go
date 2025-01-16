package main

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	_ "github.com/simplecontainer/smr/docs"
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/simplecontainer/smr/pkg/commands"
	_ "github.com/simplecontainer/smr/pkg/commands"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/helpers"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	_ "net/http/pprof"
	"os"
)

func main() {

	startup.SetFlags()

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = static.DEFAULT_LOG_LEVEL
	}

	logger.Log = logger.NewLogger(logLevel, []string{"stdout"}, []string{"stderr"})
	fmt.Println(fmt.Sprintf("logging level set to %s (override with LOG_LEVEL env variable)", logLevel))

	// Prepare configuration for the commands
	conf := configuration.NewConfig()
	conf.Environment = startup.GetEnvironmentInfo()
	startup.ReadFlags(conf)

	var db *badger.DB
	api := api.NewApi(conf, db)
	api.VersionServer = SMR_VERSION
	api.Manager.LogLevel = helpers.GetLogLevel(logLevel)

	// Run any commands before starting daemon
	commands.PreloadCommands()
	commands.Run(api)
}

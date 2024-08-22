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
	"os"
)

//	@title			Simple container manager API
//	@version		1.0
//	@description	This is a container orchestrator service.
//	@termsOfService	http://smr.qdnqn.com/terms

//	@contact.name	API Support
//	@contact.url	https://github.com/simplecontainer/smr

//	@license.name	GNU General Public License v3.0
//	@license.url	https://github.com/simplecontainer/smr/blob/main/LICENSE

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.basic	BasicAuth

//	@externalDocs.description	OpenAPI
//	@externalDocs.url			https://swagger.io/resources/open-api/

func main() {
	startup.SetFlags()
	logger.Log = logger.NewLogger()

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = static.DEFAULT_LOG_LEVEL
	}

	fmt.Println(fmt.Sprintf("logging level set to %s (override with LOG_LEVEL env variable)", logLevel))

	// Prepare configuration for the commands
	conf := configuration.NewConfig()
	conf.Environment = startup.GetEnvironmentInfo()
	startup.ReadFlags(conf)

	var db *badger.DB
	api := api.NewApi(conf, db)
	api.Manager.LogLevel = helpers.GetLogLevel(logLevel)

	// Run any commands before starting daemon
	commands.PreloadCommands()
	commands.Run(api)
}

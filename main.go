package main

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	mdns "github.com/miekg/dns"
	"github.com/spf13/viper"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	_ "smr/docs"
	"smr/pkg/api"
	"smr/pkg/commands"
	_ "smr/pkg/commands"
	"smr/pkg/config"
	"smr/pkg/logger"
	"strconv"
)

//	@title			Simple container manager API
//	@version		1.0
//	@description	This is a container orchestrator service.
//	@termsOfService	http://smr.qdnqn.com/terms

//	@contact.name	API Support
//	@contact.url	https://github.com/qdnqn/smr

//	@license.name	GNU General Public License v3.0
//	@license.url	https://github.com/qdnqn/smr/blob/main/LICENSE

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.basic	BasicAuth

//	@externalDocs.description	OpenAPI
//	@externalDocs.url			https://swagger.io/resources/open-api/

func main() {
	logger.Log = logger.NewLogger()

	conf := config.NewConfig()
	conf.ReadFlags()

	var db *badger.DB
	var err error

	if viper.GetBool("optmode") {
		// Instance of the key value store if the optmode enabled
		db, err = badger.Open(badger.DefaultOptions("/home/smr-agent/smr/smr/persistent/kv-store/badger"))
		if err != nil {
			logger.Log.Fatal(err.Error())
		}
		defer db.Close()
	}

	api := api.NewApi(conf, db)

	commands.PreloadCommands()
	commands.Run(api.Manager)

	if viper.GetBool("daemon") {
		conf.Load(api.Runtime.PROJECTDIR)

		mdns.HandleFunc(".", api.HandleDns)

		// start dns server in go routine to detach from main
		port := 53
		server := &mdns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
		go server.ListenAndServe()
		defer server.Shutdown()

		if viper.GetBool("cilium") {
			// Start cilium containers for multi host networking
		}

		api.Manager.Reconcile()
		router := gin.Default()

		v1 := router.Group("/api/v1")
		{
			operators := v1.Group("/operators")
			{
				operators.GET(":group", api.ListSupported)
				operators.GET(":group/:operator", api.RunOperators)
				operators.POST(":group/:operator", api.RunOperators)
			}

			objects := v1.Group("/")
			{
				objects.POST("apply", api.Apply)
			}

			containers := v1.Group("/")
			{
				containers.GET("ps", api.Ps)
			}

			database := v1.Group("database")
			{
				database.GET(":key", api.DatabaseGet)
				database.GET("keys", api.DatabaseGetKeys)
				database.GET("keys/:prefix", api.DatabaseGetKeysPrefix)
				database.POST(":key", api.DatabaseSet)
				database.PUT(":key", api.DatabaseSet)
			}
		}

		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		router.GET("/healthz", api.Health)

		router.Run()
	}
}

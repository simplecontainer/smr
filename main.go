package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	mdns "github.com/miekg/dns"
	_ "github.com/simplecontainer/smr/docs"
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/simplecontainer/smr/pkg/commands"
	_ "github.com/simplecontainer/smr/pkg/commands"
	"github.com/simplecontainer/smr/pkg/config"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/plugins"
	"github.com/spf13/viper"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strconv"
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

		router := gin.New()

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
				objects.POST("compare", api.Compare)
				objects.POST("delete", api.Delete)
			}

			containers := v1.Group("/")
			{
				containers.GET("ps", api.Ps)
			}

			database := v1.Group("database")
			{
				database.GET("get/:key", api.DatabaseGet)
				database.GET("keys", api.DatabaseGetKeys)
				database.GET("keys/:prefix", api.DatabaseGetKeysPrefix)
				database.POST("create/:key", api.DatabaseSet)
				database.PUT("update/:key", api.DatabaseSet)
			}

			secrets := v1.Group("secrets")
			{
				secrets.GET("get/:secret", api.SecretsGet)
				secrets.GET("keys", api.SecretsGetKeys)
				secrets.POST("create/:secret", api.SecretsSet)
				secrets.PUT("update/:secret", api.SecretsSet)
			}
		}

		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		router.GET("/healthz", api.Health)

		if viper.GetBool("daemon-secured") {
			api.Keys = keys.NewKeys("/home/smr-agent/.ssh")
			api.Manager.Keys = api.Keys

			found, err := api.Keys.GenerateIfNoKeysFound()

			if err != nil {
				panic("failed to generate or read mtls keys")
			}

			if !found {
				err := api.Keys.SaveToDirectory()

				if err != nil {
					logger.Log.Error("failed to save keys to directory", zap.String("error", err.Error()))
					os.Exit(1)
				}

				fmt.Println("Certificate is generated for the use by the smr client!")
				fmt.Println("Copy-paste it to safe location for further use - it will not be printed anymore in the logs")
				fmt.Println(api.Keys.GeneratePemBundle())
			}

			api.SetupEncryptedDatabase(api.Keys.ServerPrivateKey.Bytes()[:32])

			certPool := x509.NewCertPool()
			if ok := certPool.AppendCertsFromPEM(api.Keys.CAPem.Bytes()); !ok {
				panic("invalid cert in CA PEM")
			}

			serverTLSCert, err := tls.X509KeyPair(api.Keys.ServerCertPem.Bytes(), api.Keys.ServerPrivateKey.Bytes())
			if err != nil {
				logger.Log.Fatal("error opening certificate and key file for control connection", zap.String("error", err.Error()))
				return
			}

			tlsConfig := &tls.Config{
				ClientAuth:   tls.RequireAndVerifyClientCert,
				ClientCAs:    certPool,
				Certificates: []tls.Certificate{serverTLSCert},
			}

			server := http.Server{
				Addr:      ":1443",
				Handler:   router,
				TLSConfig: tlsConfig,
			}

			plugins.StartPlugins(api.Config.Configuration.Environment.Root, api.Manager)

			defer server.Close()
			server.ListenAndServeTLS("", "")
		} else {
			router.Run()
		}
	}
}

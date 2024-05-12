package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	mdns "github.com/miekg/dns"
	"github.com/qdnqn/smr/pkg/api"
	"github.com/qdnqn/smr/pkg/commands"
	_ "github.com/qdnqn/smr/pkg/commands"
	"github.com/qdnqn/smr/pkg/config"
	"github.com/qdnqn/smr/pkg/keys"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/spf13/viper"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"net/http"
	_ "smr/docs"
	"strconv"
	"strings"
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

		if viper.GetBool("daemon-secured") {
			mtls := keys.NewKeys("/home/smr/.ssh")

			found, err := mtls.GenerateIfNoKeysFound()

			if err != nil {
				panic("failed to generate or read mtls keys")
			}

			if !found {
				mtls.SaveToDirectory()

				fmt.Println("Certificate is generated for the use by the smr client!")
				fmt.Println("Copy-paste it to safe location for further use - it will not be printed anymore in the logs")
				fmt.Println(fmt.Sprintf("%s\n%s\n%s\n", strings.TrimSpace(mtls.ClientPrivateKey.String()), strings.TrimSpace(mtls.ClientCertPem.String()), strings.TrimSpace(mtls.CAPem.String())))
			}

			certPool := x509.NewCertPool()
			if ok := certPool.AppendCertsFromPEM(mtls.CAPem.Bytes()); !ok {
				panic("invalid cert in CA PEM")
			}

			serverTLSCert, err := tls.X509KeyPair(mtls.ServerCertPem.Bytes(), mtls.ServerPrivateKey.Bytes())
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
			defer server.Close()
			server.ListenAndServeTLS("", "")
		} else {
			router.Run()
		}
	}
}

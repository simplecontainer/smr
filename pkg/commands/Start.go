package commands

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/gin-gonic/gin"
	mdns "github.com/miekg/dns"
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/mtls"
	"github.com/simplecontainer/smr/pkg/plugins"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strconv"
)

func Start() {
	Commands = append(Commands, Command{
		name: "start",
		condition: func(*api.Api) bool {
			return true
		},
		functions: []func(*api.Api, []string){
			func(api *api.Api, args []string) {
				configFile, err := os.Open(fmt.Sprintf("%s/%s/config.yaml", api.Config.Environment.PROJECTDIR, static.CONFIGDIR))

				if err != nil {
					panic(err)
				}

				conf, err := startup.Load(configFile)

				if err != nil {
					panic(err)
				}

				api.Config = conf
				api.Config.Environment = startup.GetEnvironmentInfo()
				startup.ReadFlags(api.Config)

				api.Manager.Config = api.Config

				api.Keys = mtls.NewKeys("/home/smr-agent/.ssh/simplecontainer")
				api.Manager.Keys = api.Keys

				var found bool
				found, err = mtls.GenerateIfNoKeysFound(api.Keys, api.Config)

				if err != nil {
					panic(err)
				}

				if !found {
					err = mtls.SaveToDirectory(api.Keys)

					if err != nil {
						logger.Log.Error("failed to save keys to directory", zap.String("error", err.Error()))
						os.Exit(1)
					}

					fmt.Println("/*********************************************************************/")
					fmt.Println("/* Certificate is generated for the use by the smr client!           */")
					fmt.Println("/* It is located in the .ssh directory in home of the running user!  */")
					fmt.Println("/* cat $HOME/.ssh/simplecontainer/client.pem                         */")
					fmt.Println("/*********************************************************************/")

					mtls.GeneratePemBundle(api.Keys)
				}

				mdns.HandleFunc(".", api.HandleDns)

				port := 53
				DNSserver := &mdns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}

				go DNSserver.ListenAndServe()
				defer DNSserver.Shutdown()

				router := gin.New()

				v1 := router.Group("/api/v1")
				{
					logs := v1.Group("/logs")
					{
						//logs.GET("/", api.Agent)
						logs.GET(":kind/:group/:identifier", api.Logs)
					}

					definitions := v1.Group("/definitions")
					{
						definitions.GET("/", api.Definitions)
						definitions.GET("/:definition", api.Definition)
					}

					dns := v1.Group("/dns")
					{
						dns.GET("/", api.ListDns)
						dns.GET("/:dns", api.ListDns)
					}

					operators := v1.Group("/operators")
					{
						operators.GET(":group", api.ListSupported)
						operators.GET(":group/:operator", api.RunOperators)
						operators.POST(":group/:operator", api.RunOperators)
					}

					objects := v1.Group("/")
					{
						objects.POST("apply", api.Apply)
						objects.POST("apply/:kind", api.Apply)
						objects.POST("compare", api.Compare)
						objects.POST("delete", api.Delete)
					}

					containers := v1.Group("/")
					{
						containers.GET("ps", api.Ps)
					}

					database := v1.Group("database")
					{
						database.POST("create/:key", api.DatabaseSet)
						database.PUT("update/:key", api.DatabaseSet)
						database.GET("get/:key", api.DatabaseGet)
						database.GET("keys", api.DatabaseGetKeys)
						database.GET("keys/:prefix", api.DatabaseGetKeysPrefix)
						database.DELETE("keys/:prefix", api.DatabaseRemoveKeys)
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
				router.GET("/version", api.Version)
				router.GET("/restore", api.Restore)

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

				api.DnsCache.AddARecord(static.SMR_AGENT_DOMAIN, api.Config.Environment.AGENTIP)

				plugins.StartPlugins(api.Config.OptRoot, api.Manager)

				defer server.Close()
				server.ListenAndServeTLS("", "")
			},
		},
		depends_on: []func(*api.Api, []string){
			func(mgr *api.Api, args []string) {},
		},
	})
}

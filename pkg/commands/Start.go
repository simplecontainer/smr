package commands

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	mdns "github.com/miekg/dns"
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/kinds"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
				var conf = configuration.NewConfig()
				var err error

				conf, err = startup.Load(api.Config.Environment)

				if err != nil {
					panic(err)
				}

				api.Config = conf
				api.Manager.Config = api.Config

				api.Keys = keys.NewKeys()
				api.Manager.Keys = api.Keys

				api.User = &authentication.User{
					Username: api.Config.NodeName,
					Domain:   "localhost:1443",
				}
				api.Manager.User = api.User

				var found error

				found = api.Keys.CAExists(static.SMR_SSH_HOME, api.Config.NodeName)

				if found != nil {
					err = api.Keys.GenerateCA()

					if err != nil {
						panic("failed to generate CA")
					}

					err = api.Keys.CA.Write(static.SMR_SSH_HOME)
					if err != nil {
						panic(err)
					}
				}

				found = api.Keys.ServerExists(static.SMR_SSH_HOME, api.Config.NodeName)

				if found != nil {
					err = api.Keys.GenerateServer(api.Config.Certificates.Domains, api.Config.Certificates.IPs)

					if err != nil {
						panic(err)
					}

					err = api.Keys.GenerateClient(api.Config.Certificates.Domains, api.Config.Certificates.IPs, api.Config.NodeName)

					if err != nil {
						panic(err)
					}

					err = api.Keys.Server.Write(static.SMR_SSH_HOME, api.Config.NodeName)
					if err != nil {
						panic(err)
					}

					err = api.Keys.Clients[api.Config.NodeName].Write(static.SMR_SSH_HOME, api.Config.NodeName)
					if err != nil {
						panic(err)
					}

					fmt.Println("/*********************************************************************/")
					fmt.Println("/* Certificate is generated for the use by the smr client!           */")
					fmt.Println("/* It is located in the .ssh directory in home of the running user!  */")
					fmt.Println("/* ls $HOME/.ssh/simplecontainer                                     */")
					fmt.Println("/*********************************************************************/")

					err = api.Keys.GeneratePemBundle(static.SMR_SSH_HOME, api.Config.NodeName, api.Keys.Clients[api.Config.NodeName])

					if err != nil {
						panic(err)
					}
				}

				api.Keys.Reloader, err = keys.NewKeypairReloader(api.Keys.Server.CertificatePath, api.Keys.Server.PrivateKeyPath)
				if err != nil {
					panic(err.Error())
				}

				err = api.Keys.LoadClients(static.SMR_SSH_HOME)

				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				// Cluster information is unknown, this only enables localhost to talk to itself via https
				api.Manager.Http, err = client.GenerateHttpClients(api.Config.NodeName, api.Keys, nil)

				if err != nil {
					panic(err)
				}

				api.DnsCache = dns.New(api.Config.NodeName, api.Manager.Http, api.User)
				api.DnsCache.Client = api.Manager.Http

				api.Manager.DnsCache = api.DnsCache
				go api.DnsCache.ListenRecords()

				mdns.HandleFunc(".", api.HandleDns)

				port := 53
				DNSServer := &mdns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}

				go func() {
					err = DNSServer.ListenAndServe()
					if err != nil {
						panic(err)
					}
				}()

				defer func(DNSserver *mdns.Server) {
					err = DNSserver.Shutdown()

					if err != nil {
						logger.Log.Error(err.Error())
						return
					}
				}(DNSServer)

				router := gin.New()
				routerHttp := gin.New()

				v1 := router.Group("/api/v1")
				{
					definition := v1.Group("/attempt")
					{
						definition.POST("/:action", api.Kind)
						definition.DELETE("/:action", api.Kind)
					}

					kind := v1.Group("kind")
					{
						kind.GET("/", api.List)
						kind.GET("/:prefix/:version/:category/:kind", api.ListKind)
						kind.GET("/:prefix/:version/:category/:kind/:group", api.ListKindGroup)
						kind.GET("/:prefix/:version/:category/:kind/:group/:name", api.GetKind)
						kind.POST("/propose/:prefix/:version/:category/:kind/:group/:name", api.ProposeKind)
						kind.POST("/:prefix/:version/:category/:kind/:group/:name", api.SetKind)
						kind.PUT("/:prefix/:version/:category/:kind/:group/:name", api.SetKind)
						kind.DELETE("/:prefix/:version/:category/:kind/:group/:name", api.DeleteKind)
					}

					key := v1.Group("key")
					{
						key.POST("/propose/*key", api.ProposeKey)
						key.POST("/set/*key", api.SetKey)
						key.DELETE("/remove/*key", api.RemoveKey)
					}

					cluster := v1.Group("cluster")
					{
						cluster.GET("/", api.GetCluster)
						cluster.POST("/start", api.StartCluster)
						cluster.POST("/node", api.AddNode)
						cluster.DELETE("/node/:node", api.RemoveNode)
					}

					definitions := v1.Group("/")
					{
						definitions.POST("propose/apply", api.Propose)
						definitions.DELETE("propose/remove", api.Propose)
						definitions.GET("debug/:prefix/:version/:category/:kind/:group/:name/:follow", api.Debug)
						definitions.GET("logs/:prefix/:version/:category/:kind/:group/:name/:follow", api.Logs)
					}

					users := v1.Group("/user")
					{
						users.POST("/:username/:domain/:externalIP", api.CreateUser)
					}
				}

				debug := router.Group("/debug", func(c *gin.Context) {
					c.Next()
				})
				pprof.RouteRegister(debug, "pprof")

				router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

				router.GET("/connect", api.Health)
				router.GET("/fetch/certs", api.ExportClients)

				router.GET("/metrics", api.MetricsHandle())
				router.GET("/healthz", api.Health)
				router.GET("/version", api.Version)

				routerHttp.GET("/metrics", api.MetricsHandle())
				routerHttp.GET("/healthz", api.Health)
				routerHttp.GET("/version", api.Version)

				CAPool := x509.NewCertPool()
				CAPool.AddCert(api.Keys.CA.Certificate)

				tlsConfig := &tls.Config{
					ClientAuth: tls.RequireAndVerifyClientCert,
					ClientCAs:  CAPool,
				}

				tlsConfig.GetCertificate = api.Keys.Reloader.GetCertificateFunc()

				api.SetupEtcd()

				server := http.Server{
					Addr:      fmt.Sprintf("%s:%s", api.Config.HostPort.Host, api.Config.HostPort.Port),
					Handler:   router,
					TLSConfig: tlsConfig,
				}

				server.TLSConfig.GetCertificate = api.Keys.Reloader.GetCertificateFunc()
				_, err = api.DnsCache.AddARecord(static.SMR_NODE_DOMAIN, api.Config.Environment.NodeIP)

				if err != nil {
					panic(err)
				}

				api.KindsRegistry = kinds.BuildRegistry(api.Manager)
				api.Manager.KindsRegistry = api.KindsRegistry

				defer func(server *http.Server) {
					err = server.Close()
					if err != nil {
						return
					}
				}(&server)

				go func() {
					httpServer := http.Server{
						Handler: routerHttp,
					}

					err = httpServer.ListenAndServe()
					if err != nil {
						panic(err)
					}
				}()

				err = server.ListenAndServeTLS("", "")
				if err != nil {
					panic(err)
				}
			},
		},
		depends_on: []func(*api.Api, []string){
			func(mgr *api.Api, args []string) {},
		},
	})
}

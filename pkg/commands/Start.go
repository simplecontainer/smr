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
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/kinds"
	middleware "github.com/simplecontainer/smr/pkg/middlewares"
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
				conf, err := startup.Load(api.Config.Environment)

				if err != nil {
					panic(err)
				}

				api.Config = conf
				api.Config.Environment = startup.GetEnvironmentInfo()
				startup.ReadFlags(api.Config)

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

				var httpClient *http.Client
				var APIEndpoint string

				for username, c := range api.Keys.Clients {
					httpClient, APIEndpoint, err = client.GenerateHttpClient(api.Keys.CA, api.Keys.Clients[username])

					if err != nil {
						panic(err)
					}

					if username == api.Config.NodeName {
						APIEndpoint = "localhost"
					}

					api.Manager.Http.Append(username, &client.Client{
						Http:     httpClient,
						Username: username,
						API:      fmt.Sprintf("%s:%s", APIEndpoint, api.Config.HostPort.Port),
						Domains:  c.Certificate.DNSNames,
						IPs:      c.Certificate.IPAddresses,
					})
				}

				api.DnsCache = dns.New(api.Config.NodeName, api.Manager.Http, api.User)
				api.DnsCache.Client = api.Manager.Http

				// Listen Records from RAFT
				go api.DnsCache.ListenRecords()

				api.Manager.DnsCache = api.DnsCache

				mdns.HandleFunc(".", api.HandleDns)

				port := 53
				DNSserver := &mdns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}

				go func() {
					err = DNSserver.ListenAndServe()
					if err != nil {
						panic(err)
					}
				}()

				defer func(DNSserver *mdns.Server) {
					err = DNSserver.Shutdown()
					if err != nil {
						return
					}
				}(DNSserver)

				router := gin.New()
				router.Use(middleware.TLSAuth())
				router.Use(api.ClusterCheck())

				v1 := router.Group("/api/v1")
				{
					database := v1.Group("database")
					{
						database.POST("create/*key", api.DatabaseSet)
						database.PUT("update/*key", api.DatabaseSet)
						database.POST("propose/:type/*key", api.ProposeDatabase)
						database.PUT("propose/:type/*key", api.ProposeDatabase)
						database.GET("get/*key", api.DatabaseGet)
						database.GET("keys", api.DatabaseGetKeys)
						database.GET("keys/*prefix", api.DatabaseGetKeysPrefix)
						database.DELETE("keys/*prefix", api.DatabaseRemoveKeys)
					}

					cluster := v1.Group("cluster")
					{
						cluster.GET("/", api.GetCluster)
						cluster.POST("/start", api.StartCluster)
						cluster.POST("/node", api.AddNode)
						cluster.DELETE("/node/:node", api.RemoveNode)
					}

					kinds := v1.Group("/")
					{
						kinds.POST("apply", api.Apply)
						kinds.POST("propose/apply", api.ProposeObject)
						kinds.DELETE("propose/remove", api.ProposeObject)
						kinds.POST("compare", api.Compare)
						kinds.DELETE("delete", api.Delete)
						kinds.GET("debug/:kind/:group/:identifier/:follow", api.Debug)
						kinds.GET("logs/:group/:identifier/:follow", api.Logs)
					}

					operators := v1.Group("/control")
					{
						operators.GET(":kind", api.ListSupported)
						operators.GET(":kind/:operation", api.RunControl)
						operators.GET(":kind/:operation/:group/:name", api.RunControl)
						operators.POST(":kind/:operation/:group/:name", api.RunControl)
						operators.PUT(":kind/:operation/:group/:name", api.RunControl)
						operators.DELETE(":kind/:operation/:group/:name", api.RunControl)
					}

					secrets := v1.Group("secrets")
					{
						secrets.POST("create/*key", api.SecretsSet)
						secrets.PUT("update/*key", api.SecretsSet)
						secrets.POST("propose/:type/*key", api.ProposeSecrets)
						secrets.PUT("propose/:type/*key", api.ProposeSecrets)
						secrets.GET("get/*key", api.SecretsGet)
						secrets.GET("keys", api.SecretsGet)
						secrets.GET("keys/*prefix", api.SecretsGetKeysPrefix)
						secrets.DELETE("keys/*prefix", api.SecretsRemoveKeys)
					}

					containers := v1.Group("/")
					{
						containers.GET("ps", api.Ps)
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

				router.GET("/metrics", api.Metrics())
				router.GET("/healthz", api.Health)
				router.GET("/version", api.Version)

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
				_, err = api.DnsCache.AddARecord(static.SMR_AGENT_DOMAIN, api.Config.Environment.AGENTIP)

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

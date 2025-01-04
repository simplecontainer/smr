package commands

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
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
					Username: api.Config.Node,
					Domain:   "localhost:1443",
				}
				api.Manager.User = api.User

				var found error

				found = api.Keys.CAExists(static.SMR_SSH_HOME, api.Config.Node)

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

				found = api.Keys.ServerExists(static.SMR_SSH_HOME, api.Config.Node)

				if found != nil {
					err = api.Keys.GenerateServer(api.Config.Certificates.Domains, api.Config.Certificates.IPs)

					if err != nil {
						panic(err)
					}

					err = api.Keys.GenerateClient(api.Config.Certificates.Domains, api.Config.Certificates.IPs, api.Config.Node)

					if err != nil {
						panic(err)
					}

					err = api.Keys.Server.Write(static.SMR_SSH_HOME, api.Config.Node)
					if err != nil {
						panic(err)
					}

					err = api.Keys.Clients[api.Config.Node].Write(static.SMR_SSH_HOME, api.Config.Node)
					if err != nil {
						panic(err)
					}

					fmt.Println("/*********************************************************************/")
					fmt.Println("/* Certificate is generated for the use by the smr client!           */")
					fmt.Println("/* It is located in the .ssh directory in home of the running user!  */")
					fmt.Println("/* ls $HOME/.ssh/simplecontainer                                     */")
					fmt.Println("/*********************************************************************/")

					err = api.Keys.GeneratePemBundle(static.SMR_SSH_HOME, api.Config.Node, api.Keys.Clients[api.Config.Node])

					if err != nil {
						panic(err)
					}
				}

				api.Keys.Reloader, err = keys.NewKeypairReloader(api.Keys.Server.CertificatePath, api.Keys.Server.PrivateKeyPath)
				if err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
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

					if username == api.Config.Node {
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

				api.DnsCache = dns.New(api.Config.Node, api.Manager.Http, api.User)
				api.DnsCache.Client = api.Manager.Http

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
						database.POST("propose/*key", api.Propose)
						database.PUT("propose/*key", api.Propose)
						database.GET("get/*key", api.DatabaseGet)
						database.GET("keys", api.DatabaseGetKeys)
						database.GET("keys/*prefix", api.DatabaseGetKeysPrefix)
						database.DELETE("keys/*prefix", api.DatabaseRemoveKeys)
					}

					definitions := v1.Group("/definitions")
					{
						definitions.GET("/", api.Definitions)
						definitions.GET("/:definition", api.Definition)
					}

					kinds := v1.Group("/")
					{
						kinds.POST("apply", api.Apply)
						kinds.POST("apply/:kind", api.Apply)
						kinds.POST("apply/:kind/:agent", api.Apply)
						kinds.POST("compare", api.Compare)
						kinds.POST("delete", api.Delete)
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
						secrets.POST("create/*key", api.DatabaseSet)
						secrets.PUT("update/*key", api.DatabaseSet)
						secrets.POST("propose/*key", api.Propose)
						secrets.PUT("propose/*key", api.Propose)
						secrets.GET("get/*key", api.DatabaseGetBase64)
						secrets.GET("keys", api.DatabaseGetKeys)
						secrets.GET("keys/*prefix", api.DatabaseGetKeysPrefix)
						secrets.DELETE("keys/*prefix", api.DatabaseRemoveKeys)
					}

					containers := v1.Group("/")
					{
						containers.GET("ps", api.Ps)
					}

					logs := v1.Group("/logs")
					{
						//logs.GET("/", api.Node)
						logs.GET(":kind/:group/:identifier", api.Logs)
					}

					dns := v1.Group("/dns")
					{
						dns.GET("/", api.ListDns)
						dns.GET("/:dns", api.ListDns)
					}

					users := v1.Group("/user")
					{
						users.POST("/:username/:domain/:externalIP", api.CreateUser)
					}

					etcd := v1.Group("etcd")
					{
						etcd.PUT("/etcd/update/*key", api.EtcdPut)
					}
				}

				router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

				router.GET("/ca", api.CA)
				router.GET("/connect", api.Health)
				router.GET("/restore", api.Restore)
				router.GET("/healthz", api.Health)
				router.GET("/version", api.Version)
				router.GET("/metrics", api.Metrics())

				router.GET("/cluster", api.GetCluster)
				router.POST("/cluster/start", api.StartCluster)
				router.POST("/cluster/node", api.AddNode)
				router.DELETE("/cluster/node/:node", api.RemoveNode)

				CAPool := x509.NewCertPool()
				CAPool.AddCert(api.Keys.CA.Certificate)

				tlsConfig := &tls.Config{
					ClientAuth: tls.RequireAndVerifyClientCert,
					ClientCAs:  CAPool,
				}

				tlsConfig.GetCertificate = api.Keys.Reloader.GetCertificateFunc()

				api.SetupEncryptedDatabase(api.Keys.Server.PrivateKeyBytes[:32])

				server := http.Server{
					Addr:      fmt.Sprintf("%s:%s", api.Config.HostPort.Host, api.Config.HostPort.Port),
					Handler:   router,
					TLSConfig: tlsConfig,
				}

				server.TLSConfig.GetCertificate = api.Keys.Reloader.GetCertificateFunc()

				api.DnsCache.AddARecord(static.SMR_AGENT_DOMAIN, api.Config.Environment.AGENTIP)

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

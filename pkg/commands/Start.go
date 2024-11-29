package commands

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/gin-gonic/gin"
	mdns "github.com/miekg/dns"
	"github.com/simplecontainer/smr/pkg/api"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/helpers"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/kinds"
	"github.com/simplecontainer/smr/pkg/logger"
	middleware "github.com/simplecontainer/smr/pkg/middlewares"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strconv"
	"strings"
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

				api.KindsRegistry = kinds.BuildRegistry(api.Manager)
				api.Manager.KindsRegistry = api.KindsRegistry

				var found error
				found = api.Keys.Exists(static.SMR_SSH_HOME, "root")

				if found != nil {
					err = api.Keys.Generate(
						append([]string{"localhost", fmt.Sprintf("smr-agent.%s", static.SMR_LOCAL_DOMAIN)}, strings.FieldsFunc(api.Config.Domain, helpers.SplitClean)...),
						append([]string{"127.0.0.1"}, strings.FieldsFunc(api.Config.ExternalIP, helpers.SplitClean)...),
					)

					if err != nil {
						logger.Log.Error(err.Error())
						os.Exit(1)
					}

					fmt.Println("/*********************************************************************/")
					fmt.Println("/* Certificate is generated for the use by the smr client!           */")
					fmt.Println("/* It is located in the .ssh directory in home of the running user!  */")
					fmt.Println("/* cat $HOME/.ssh/simplecontainer/root.pem                           */")
					fmt.Println("/*********************************************************************/")

					err = api.Keys.CA.Write(static.SMR_SSH_HOME)
					if err != nil {
						panic(err)
					}

					err = api.Keys.Server.Write(static.SMR_SSH_HOME)
					if err != nil {
						panic(err)
					}

					err = api.Keys.Clients["root"].Write(static.SMR_SSH_HOME, "root")
					if err != nil {
						panic(err)
					}

					err = api.Keys.GeneratePemBundle(static.SMR_SSH_HOME, "root", api.Keys.Clients["root"])

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

				for username, c := range api.Keys.Clients {
					var httpClient *http.Client
					httpClient, err = client.GenerateHttpClient(api.Keys.CA, api.Keys.Clients["root"])
					if err != nil {
						panic(err)
					}

					api.Manager.Http.Append(username, &client.Client{
						Http: httpClient,
						API:  fmt.Sprintf("%s:1443", c.Certificate.DNSNames[0]),
					})
				}

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

				v1 := router.Group("/api/v1")
				{
					users := v1.Group("/user")
					{
						users.POST("/:username/:domain/:externalIP", api.CreateUser)
					}

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

				router.POST("/cluster/start", api.StartCluster)
				router.POST("/cluster/node", api.AddNode)
				router.DELETE("/cluster/node/:node", api.RemoveNode)
				router.GET("/cluster", api.GetCluster)
				router.PUT("/etcd/update/*key", api.EtcdPut)

				CAPool := x509.NewCertPool()
				CAPool.AddCert(api.Keys.CA.Certificate)

				var PEMCertificate []byte
				var PEMPrivateKey []byte

				PEMCertificate, err = keys.PEMEncode(keys.CERTIFICATE, api.Keys.Server.CertificateBytes)
				PEMPrivateKey, err = keys.PEMEncode(keys.PRIVATE_KEY, api.Keys.Server.PrivateKeyBytes)

				serverTLSCert, err := tls.X509KeyPair(PEMCertificate, PEMPrivateKey)
				if err != nil {
					logger.Log.Fatal("error opening certificate and key file for control connection", zap.String("error", err.Error()))
					return
				}

				tlsConfig := &tls.Config{
					ClientAuth:   tls.RequireAndVerifyClientCert,
					ClientCAs:    CAPool,
					Certificates: []tls.Certificate{serverTLSCert},
				}

				api.SetupEncryptedDatabase(api.Keys.Server.PrivateKeyBytes[:32])

				server := http.Server{
					Addr:      ":1443",
					Handler:   router,
					TLSConfig: tlsConfig,
				}

				server.TLSConfig.GetCertificate = api.Keys.Reloader.GetCertificateFunc()

				api.DnsCache.AddARecord(static.SMR_AGENT_DOMAIN, api.Config.Environment.AGENTIP)

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

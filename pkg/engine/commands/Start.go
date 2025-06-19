package commands

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/gin-gonic/gin"
	mdns "github.com/miekg/dns"
	"github.com/simplecontainer/smr/pkg/api/middlewares"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/command"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/kinds"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"strconv"
	"time"
)

func Start() {
	Commands = append(Commands,
		command.NewBuilder().Parent("smr").Name("start").Function(cmdStart).Flags(cmdStartFlags).Build(),
	)
}

func cmdStart(api iapi.Api, cli *client.Client, args []string) {
	var conf = configuration.NewConfig()
	var err error

	conf, err = startup.Load(api.GetConfig().Environment.Container)

	if err != nil {
		panic(err)
	}

	api.SetConfig(conf)
	api.GetManager().Config = api.GetConfig()

	api.SetKeys(keys.NewKeys())
	api.GetManager().Keys = api.GetKeys()

	api.SetUser(&authentication.User{
		Username: api.GetConfig().NodeName,
		Domain:   "localhost:1443",
	})

	api.GetManager().User = api.GetUser()
	api.GetVersion().Image = api.GetConfig().NodeImage

	var found error

	found = api.GetKeys().CAExists(static.SMR_SSH_HOME, api.GetConfig().NodeName)

	if found != nil {
		err = api.GetKeys().GenerateCA()

		if err != nil {
			panic("failed to generate CA")
		}

		err = api.GetKeys().CA.Write(static.SMR_SSH_HOME)
		if err != nil {
			panic(err)
		}
	}

	found = api.GetKeys().ServerExists(static.SMR_SSH_HOME, api.GetConfig().NodeName)

	if found != nil {
		err = api.GetKeys().GenerateServer(api.GetConfig().Certificates.Domains, api.GetConfig().Certificates.IPs)

		if err != nil {
			panic(err)
		}

		err = api.GetKeys().GenerateClient(api.GetConfig().Certificates.Domains, api.GetConfig().Certificates.IPs, api.GetConfig().NodeName)

		if err != nil {
			panic(err)
		}

		err = api.GetKeys().Server.Write(static.SMR_SSH_HOME, api.GetConfig().NodeName)
		if err != nil {
			panic(err)
		}

		err = api.GetKeys().Clients[api.GetConfig().NodeName].Write(static.SMR_SSH_HOME, api.GetConfig().NodeName)
		if err != nil {
			panic(err)
		}

		fmt.Println("/*********************************************************************/")
		fmt.Println("/* Certificate is generated for the use by the smr client!           */")
		fmt.Println("/* It is located in the .ssh directory in home of the running user!  */")
		fmt.Println("/* ls $HOME/.ssh/simplecontainer                                     */")
		fmt.Println("/*********************************************************************/")

		err = api.GetKeys().GeneratePemBundle(static.SMR_SSH_HOME, api.GetConfig().NodeName, api.GetKeys().Clients[api.GetConfig().NodeName])

		if err != nil {
			panic(err)
		}
	}

	api.GetKeys().Reloader, err = keys.NewKeypairReloader(api.GetKeys().Server.CertificatePath, api.GetKeys().Server.PrivateKeyPath)
	if err != nil {
		panic(err.Error())
	}

	err = api.GetKeys().LoadClients(static.SMR_SSH_HOME)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Cluster information is unknown, this only enables localhost to talk to itself via https
	api.GetManager().Http, err = clients.GenerateHttpClients(api.GetKeys(), api.GetConfig().HostPort, nil)

	if err != nil {
		panic(err)
	}

	api.SetDnsCache(dns.New(api.GetConfig().NodeName, api.GetManager().Http, api.GetUser()))
	api.GetDnsCache().Client = api.GetManager().Http

	api.GetManager().DnsCache = api.GetDnsCache()
	go api.GetDnsCache().ListenRecords()

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

	router.Use(middlewares.CORS())

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
			kind.POST("/compare/:prefix/:version/:category/:kind/:group/:name", api.CompareKind)
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
			cluster.GET("/", api.StatusCluster)
			cluster.POST("/start", api.StartCluster)
			cluster.POST("/control", api.Control)
			cluster.GET("/nodes", api.Nodes)
			cluster.GET("/node/:id", api.GetNode)
			cluster.GET("/node/version/:id", api.GetNodeVersion)
			cluster.POST("/node", api.AddNode)
			cluster.DELETE("/node/:node", api.RemoveNode)
		}

		definitions := v1.Group("/")
		{
			definitions.POST("propose/:action", api.Propose)
			definitions.DELETE("propose/:action", api.Propose)
			definitions.GET("debug/:prefix/:version/:category/:kind/:group/:name/:which/:follow", api.Debug)
			definitions.GET("logs/:prefix/:version/:category/:kind/:group/:name/:which/:follow", api.Logs)
			definitions.GET("exec/:prefix/:version/:kind/:containers/:group/:name/:interactive", api.Exec)
		}

		users := v1.Group("/user")
		{
			users.POST("/:username/:domain/:externalIP", api.CreateUser)
		}
	}

	router.GET("/connect", api.Health)
	router.GET("/fetch/certs", api.ExportClients)

	router.GET("/metrics", api.MetricsHandle())
	router.GET("/healthz", api.Health)
	router.GET("/version", api.DisplayVersion)
	router.GET("/events", api.Events)

	//debug := routerHttp.Group("/debug", func(c *gin.Context) {
	//	c.Next()
	//})
	//pprof.RouteRegister(debug, "pprof")

	routerHttp.GET("/metrics", api.MetricsHandle())
	routerHttp.GET("/healthz", api.Health)
	routerHttp.GET("/version", api.DisplayVersion)

	CAPool := x509.NewCertPool()
	CAPool.AddCert(api.GetKeys().CA.Certificate)

	tlsConfig := &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  CAPool,
	}

	tlsConfig.GetCertificate = api.GetKeys().Reloader.GetCertificateFunc()

	api.SetupEtcd()

	server := http.Server{
		Addr:         fmt.Sprintf("%s:%s", api.GetConfig().HostPort.Host, api.GetConfig().HostPort.Port),
		Handler:      router,
		TLSConfig:    tlsConfig,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	server.TLSConfig.GetCertificate = api.GetKeys().Reloader.GetCertificateFunc()
	_, err = api.GetDnsCache().AddARecord(static.SMR_NODE_DOMAIN, api.GetConfig().Environment.Container.NodeIP)

	if err != nil {
		panic(err)
	}

	api.SetKindsRegistry(kinds.BuildRegistry(api.GetManager()))
	api.GetManager().KindsRegistry = api.GetKindsRegistry()

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
}

func cmdStartFlags(cmd *cobra.Command) {
	cmd.Flags().String("flannel.backend", "wireguard", "Flannel backend: vxlan, wireguard")
	cmd.Flags().String("flannel.cidr", "10.10.0.0/16", "Flannel overlay network CIDR")
	cmd.Flags().String("flannel.iface", "", "Network interface for flannel to use, if ommited default gateway will be used")
}

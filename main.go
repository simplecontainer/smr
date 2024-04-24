package main

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	mdns "github.com/miekg/dns"
	"github.com/spf13/viper"
	"smr/pkg/api"
	"smr/pkg/commands"
	_ "smr/pkg/commands"
	"smr/pkg/config"
	"smr/pkg/logger"
	"strconv"
)

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

		// start server in go routine to detach from main
		port := 53
		server := &mdns.Server{Addr: ":" + strconv.Itoa(port), Net: "udp"}
		go server.ListenAndServe()
		defer server.Shutdown()

		api.Manager.Reconcile()
		router := gin.Default()

		// System
		router.GET("/healthz", api.Health)

		// Operators
		router.GET("/operators/:group/", api.ListSupported)
		router.GET("/operators/:group/:operator", api.RunOperators)
		router.POST("/operators/:group/:operator", api.RunOperators)

		// Containers
		router.POST("/apply", api.Apply)
		router.GET("/ps", api.Ps)

		// Definition
		// router.GET("/definition", mgr.Api.List)
		// router.GET("/definition/:definitionName", mgr.Api.List)

		// Database

		router.GET("/database/:key", api.DatabaseGet)
		router.POST("/database/:key", api.DatabaseSet)
		router.PUT("/database/:key", api.DatabaseSet)

		router.Run()
	}
}

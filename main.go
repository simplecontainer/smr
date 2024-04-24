package main

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"smr/pkg/api"
	"smr/pkg/commands"
	_ "smr/pkg/commands"
	"smr/pkg/config"
	"smr/pkg/logger"
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

		go api.Manager.Reconcile()
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

package agent

import (
	"context"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contexts"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/spf13/viper"
)

func Events() {
	environment := configuration.NewEnvironment(configuration.WithHostConfig())
	conf, err := startup.Load(environment)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	cli := client.New(conf, environment.NodeDirectory)
	cli.Context, err = contexts.LoadActive(contexts.DefaultConfig(environment.NodeDirectory))

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	err = cli.Events(ctx, cancel, viper.GetString("wait"), "", cli.Tracker)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}
}

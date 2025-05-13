package agent

import (
	"context"
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts/icontrol"
	"github.com/simplecontainer/smr/pkg/startup"
)

func Restart(batch icontrol.Batch) {
	environment := configuration.NewEnvironment(configuration.WithHostConfig())
	conf, err := startup.Load(environment)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	cli := client.New(conf, environment.NodeDirectory)
	cli.Context, err = client.LoadActive(client.DefaultConfig(environment.NodeDirectory))

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	response, err := batch.Apply(ctx, cli)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Println(response.Explanation)
}

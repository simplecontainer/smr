package agent

import (
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/startup"
)

func Export(api string) {
	environment := configuration.NewEnvironment(configuration.WithHostConfig())
	conf, err := startup.Load(environment)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	cli := client.New(conf, environment.NodeDirectory)

	encrypted, key, err := cli.Manager.ExportContext("", api)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Println(encrypted, key)
}

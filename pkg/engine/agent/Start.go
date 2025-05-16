package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts/icontrol"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	"net"
	"net/http"
)

func Start(batch icontrol.Batch) {
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

	var data []byte
	data, err = json.Marshal(batch)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	_, port, err := net.SplitHostPort(conf.Ports.Control)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	response := network.Send(cli.Context.GetClient(), fmt.Sprintf("%s/api/v1/cluster/start", fmt.Sprintf("https://localhost:%s", port)), http.MethodPost, data)

	if response.HttpStatus == http.StatusOK || response.ErrorExplanation == static.CLUSTER_STARTED {
		if response.HttpStatus == http.StatusOK {
			fmt.Println(response.Explanation)
		} else {
			helpers.PrintAndExit(errors.New(response.ErrorExplanation), 1)
		}
	} else {
		helpers.PrintAndExit(errors.New(response.ErrorExplanation), 1)
	}
}

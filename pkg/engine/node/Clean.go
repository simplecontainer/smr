package node

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/internal/definitions"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
)

func Clean() {
	conf, err := startup.Load(configuration.NewEnvironment(configuration.WithHostConfig()))

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	definition, err := definitions.Node(conf.NodeName, conf, nil, nil)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	var container platforms.IPlatform

	switch conf.Platform {
	case static.PLATFORM_DOCKER:
		if err = docker.IsDaemonRunning(); err != nil {
			helpers.PrintAndExit(err, 1)
		}

		container, err = docker.New(conf.NodeName, definition)
		break
	default:
		helpers.PrintAndExit(errors.New("platform unknown"), 1)
	}

	defer func() {
		err = container.Delete()

		if err != nil {
			helpers.PrintAndExit(err, 1)
		}

		fmt.Println("node deleted")
	}()

	err = container.Stop(static.SIGTERM)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	err = container.Wait("removed")

	fmt.Println("node stopped")
}

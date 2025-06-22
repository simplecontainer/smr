package node

import (
	"errors"
	"github.com/simplecontainer/smr/internal/definitions"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
)

func Container() (platforms.IPlatform, error) {
	conf, err := startup.Load(configuration.NewEnvironment(configuration.WithHostConfig()))

	if err != nil {
		return nil, err
	}

	definition, err := definitions.Node(conf.NodeName, conf, nil, nil)

	if err != nil {
		return nil, errors.New("platform unknown")
	}

	var container platforms.IPlatform

	switch conf.Platform {
	case static.PLATFORM_DOCKER:
		if err = docker.IsDaemonRunning(); err != nil {
			return nil, errors.New("platform unknown")
		}

		container, err = docker.New(conf.NodeName, definition)
		break
	default:
		return nil, errors.New("platform unknown")
	}

	return container, nil
}

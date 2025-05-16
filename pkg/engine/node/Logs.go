package node

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/mattn/go-shellwords"
	"github.com/simplecontainer/smr/internal/definitions"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	"io"
	"os"
	"strings"
)

func Logs() {
	environment := configuration.NewEnvironment(configuration.WithHostConfig())
	conf, err := startup.Load(environment)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	entrypointParsed, _ := shellwords.Parse("")
	containerArgsParsed, _ := shellwords.Parse("")

	definition, err := definitions.Node(conf.NodeName, conf, entrypointParsed, containerArgsParsed)

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
		helpers.PrintAndExit(errors.New("platform not supported"), 1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logs, err := container.Logs(ctx, false)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	defer logs.Close()

	var outBuilder, errBuilder strings.Builder

	_, err = stdcopy.StdCopy(&outBuilder, &errBuilder, logs)
	if err != nil && err != io.EOF {
		helpers.PrintAndExit(err, 1)
	}

	if out := outBuilder.String(); out != "" {
		fmt.Print(out)
	}
	if errOut := errBuilder.String(); errOut != "" {
		fmt.Fprint(os.Stderr, errOut)
	}

	return
}

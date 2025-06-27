package node

import (
	"context"
	"errors"
	"fmt"
	"github.com/mattn/go-shellwords"
	"github.com/simplecontainer/smr/internal/definitions"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/startup"
	"github.com/simplecontainer/smr/pkg/static"
	"net"
	"os"
	"time"
)

func Start(entrypoint string, args string) {
	environment := configuration.NewEnvironment(configuration.WithHostConfig())
	conf, err := startup.Load(environment)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	entrypointParsed, _ := shellwords.Parse(entrypoint)
	containerArgsParsed, _ := shellwords.Parse(args)

	definition, err := definitions.Node(conf.NodeName, conf, entrypointParsed, containerArgsParsed)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	var container platforms.IPlatform

	switch conf.Platform {
	case static.PLATFORM_DOCKER:
		if _, err = docker.IsDaemonRunning(); err != nil {
			helpers.PrintAndExit(err, 1)
		}

		container, err = docker.New(conf.NodeName, definition)
		break
	default:
		helpers.PrintAndExit(errors.New("platform not supported"), 1)
	}

	state, err := container.GetState()

	switch state.State {
	case "running":
		helpers.PrintAndExit(errors.New("container is already running"), 1)
		break
	default:
		err = container.Delete()

		if err != nil {
			helpers.PrintAndExit(err, 1)
		}
	}

	fmt.Println(fmt.Sprintf("starting node with the user: %s", definition.Spec.User))
	err = container.Run()

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Println("node started")
}

func SetupAccess() {
	environment := configuration.NewEnvironment(configuration.WithHostConfig())
	conf, err := startup.Load(environment)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err = helpers.WaitForFileToAppear(ctx, fmt.Sprintf("%s/.ssh/%s.pem", environment.NodeDirectory, conf.NodeName), 500*time.Millisecond)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	var bundle []byte
	bundle, err = os.ReadFile(fmt.Sprintf("%s/.ssh/%s.pem", environment.NodeDirectory, conf.NodeName))

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	_, port, err := net.SplitHostPort(conf.Ports.Control)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	var credentials *client.Credentials
	credentials, err = client.BundleToCredentials(bundle)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	cli := client.New(conf, environment.NodeDirectory)
	cli.Context, err = cli.Manager.CreateContext(conf.NodeName, fmt.Sprintf("https://localhost:%s", port), credentials)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	err = cli.Context.Connect(context.Background(), true)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	if err = cli.Context.Save(); err != nil {
		helpers.PrintAndExit(err, 1)
	}

	err = cli.Manager.SetActive(cli.Context.Name)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Println("context saved")
}

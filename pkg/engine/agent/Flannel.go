package agent

import (
	"context"
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contexts"
	"github.com/simplecontainer/smr/pkg/flannel"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/startup"
	"go.uber.org/zap"
	"os"
)

func Flannel() {
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

	fmt.Println("trying to run flannel if not running")
	err = helpers.AcquireLock("/var/run/flannel/flannel.lock")

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	defer func() {
		err = helpers.ReleaseLock("/var/run/flannel/flannel.lock")
		if err != nil {
			logger.Log.Error("failed to clear lock /var/run/flannel/flannel.lock - do it manually", zap.Error(err))
		}
	}()

	err = os.WriteFile("/var/run/flannel.pid", []byte(fmt.Sprintf("%d", os.Getpid())), 0644)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)

	go func() {
		logger.Log.Info("starting flannel")
		err = flannel.Run(ctx, cancel, cli, conf)

		if err != nil {
			logger.Log.Error("flannel error:", zap.Error(err))
		} else {
			logger.Log.Info("flannel exited - bye bye")
		}

		done <- err
	}()

	select {
	case <-ctx.Done():
		logger.Log.Info("agent exited: context canceled")
	case err = <-done:
		if err != nil {
			logger.Log.Error("agent exited: flannel exited with error", zap.Error(err))
		} else {
			logger.Log.Info("agent exited: flannel exited with nil")
		}
	}
}

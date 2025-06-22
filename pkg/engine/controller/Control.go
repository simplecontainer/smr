package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/control"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/startup"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"os"
	"time"
)

func Control(api iapi.Api) {
	if os.Getuid() < 1000 {
		panic("control process can only be started with 1000 and greater UIDs")
	}

	lockfile := fmt.Sprintf("/var/run/user/%d/control.lock", os.Getuid())
	pidfile := fmt.Sprintf("/var/run/user/%d/control.pid", os.Getuid())

	defer func() {
		err := helpers.ReleaseLock(lockfile)
		if err != nil {
			logger.Log.Error("failed to clear lock file - do it manually", zap.Error(err))
		}
	}()

	environment := configuration.NewEnvironment(configuration.WithHostConfig())
	conf, err := startup.Load(environment)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	fmt.Println("trying to run control watcher if not running")
	err = helpers.AcquireLock(lockfile)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	defer func() {
		err = helpers.ReleaseLock(lockfile)

		if err != nil {
			logger.Log.Error("failed to clear lock file - do it manually", zap.String("file", lockfile), zap.Error(err))
		}
	}()

	err = os.WriteFile(pidfile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{fmt.Sprintf("localhost:%s", conf.Ports.Etcd)},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		helpers.PrintAndExit(err, 1)
	}
	defer cli.Close()

	logger.Log.Info("listening for control events...")
	watchCh := cli.Watch(context.Background(), "/smr/control/", clientv3.WithPrefix())

	for watchResp := range watchCh {
		for _, event := range watchResp.Events {
			if event.Type != mvccpb.PUT {
				continue
			}

			logger.Log.Info("new control event received")

			b := control.NewCommandBatch()

			err = json.Unmarshal(event.Kv.Value, b)

			if err != nil {
				logger.Log.Error("failed to unmarshal control", zap.Error(err))
				continue
			}

			for _, cmd := range b.GetCommands() {
				err = cmd.Agent(api, cmd.Data())

				if err != nil {
					logger.Log.Error("error executing control on agent side", zap.Error(err))
				}
			}
		}
	}
}

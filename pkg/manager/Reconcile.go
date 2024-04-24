package manager

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"smr/pkg/container"
	"smr/pkg/logger"
	"time"
)

func (mgr *Manager) Reconcile() {
	mgr.ListenEvents()
}

func (mgr *Manager) ListenEvents() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	cEvents, cErr := cli.Events(ctx, types.EventsOptions{})

	for {
		select {
		case err := <-cErr:
			fmt.Println(err)
		case msg := <-cEvents:
			mgr.HandleEvent(msg)
		}
	}
}

func (mgr *Manager) HandleEvent(event events.Message) {
	if event.Status == "kill" {
		logger.Log.Info(fmt.Sprintf("Detecting the kill of the %s", event.Actor.Attributes["name"]))
	}

	if event.Status == "stop" {
		logger.Log.Info(fmt.Sprintf("Detecting the stop of the %s", event.Actor.Attributes["name"]))
	}

	if event.Status == "die" {
		logger.Log.Info(fmt.Sprintf("Detecting the die of the %s", event.Actor.Attributes["name"]))

		container := mgr.Registry.Find(event.Actor.Attributes["group"], event.Actor.Attributes["name"])
		go mgr.HandleReconcile(container)
	}
}

func (mgr *Manager) HandleReconcile(container *container.Container) {
	mgr.Registry.BackOffTracking(container.Static.Group, container.Static.Name)

	for {
		logger.Log.Info(fmt.Sprintf("Trying to reconcile %s to the defined state", container.Static.Name))

		if mgr.Registry.BackOffTracker[container.Static.Group][container.Static.Name] > 5 {
			logger.Log.Error(fmt.Sprintf("%s container is backoff restarting", container.Static.Name))
			mgr.Registry.BackOffReset(container.Static.Group, container.Static.Name)
			return
		}

		if container.Delete() {
			var err error

			logger.Log.Info(fmt.Sprintf("trying to reconcile %s", container.Static.Name))

			mgr.Prepare(container)
			_, err = container.Run(mgr.Runtime, mgr.Badger)

			if err != nil {
				logger.Log.Error(fmt.Sprintf("Failed to reconcile %s to the defined state", container.Static.Name))
			} else {
				logger.Log.Info(fmt.Sprintf("Reconcile of %s succeed", container.Static.Name))
				return
			}
		} else {
			logger.Log.Error(fmt.Sprintf("Failed to delete %s container so that reconciliation can begin", container.Static.Name))
		}

		waitingTime := time.Duration(mgr.Registry.BackOffTracker[container.Static.Group][container.Static.Name]) * time.Second
		time.Sleep(waitingTime)
	}
}

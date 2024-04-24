package manager

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
	"smr/pkg/container"
	"smr/pkg/logger"
	"time"
)

func (mgr *Manager) Reconcile() {
	go mgr.Reconciler.ListenQueue(mgr.Registry, mgr.Runtime, mgr.Badger, mgr.DnsCache)
	go mgr.Reconciler.ListenEvents(mgr.Registry)
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

func (mgr *Manager) ConfigChangeEmit(group string, identifier string) {
	/*if mgr.Registry.Containers[group] != nil {
		if identifier == "*" {
			for identifierFromRegistry, _ := range mgr.Registry.Containers[group] {
				mgr.Registry.Containers[group][identifierFromRegistry].Status.Reconciling = true
				mgr.HandleReconcile(mgr.Registry.Containers[group][identifierFromRegistry])
			}
		}	else {
			if mgr.Registry.Containers[group][identifier] != nil {
				mgr.Registry.Containers[group][identifier].Status.Reconciling = true
				mgr.HandleReconcile(mgr.Registry.Containers[group][identifier])
			}
		}
	}

	*/
}

func (mgr *Manager) ResourceChangeEmit(group string, identifier string) {
	/*
		if mgr.Registry.Containers[group] != nil {
			if identifier == "*" {
				for identifierFromRegistry, _ := range mgr.Registry.Containers[group] {
					if mgr.Registry.Containers[group][identifierFromRegistry].Status.Running {
						mgr.Registry.Containers[group][identifierFromRegistry].Status.Reconciling = true
						mgr.HandleReconcile(mgr.Registry.Containers[group][identifierFromRegistry])
					}
				}
			}	else {
				if mgr.Registry.Containers[group][identifier] != nil {
					if mgr.Registry.Containers[group][identifier].Status.Running {
						mgr.Registry.Containers[group][identifier].Status.Reconciling = true
						mgr.HandleReconcile(mgr.Registry.Containers[group][identifier])
					}
				}
			}
		}
	*/
}

func (mgr *Manager) HandleEvent(event events.Message) {
	if event.Status == "start" {
		container := mgr.Registry.Find(event.Actor.Attributes["group"], event.Actor.Attributes["name"])

		if mgr.SmrManaged(container) {
			container.Status.Running = true
			container.Status.Reconciling = false
		}
	}

	if event.Status == "kill" {
		logger.Log.Info(fmt.Sprintf("Detecting the kill of the %s", event.Actor.Attributes["name"]))
	}

	if event.Status == "stop" {
		logger.Log.Info(fmt.Sprintf("Detecting the stop of the %s", event.Actor.Attributes["name"]))
	}

	if event.Status == "die" {
		container := mgr.Registry.Find(event.Actor.Attributes["group"], event.Actor.Attributes["name"])
		container.Status.Running = false

		if mgr.EligibleForReconcile(container) {
			logger.Log.Info(fmt.Sprintf("Detecting the die of the %s", event.Actor.Attributes["name"]))
			go mgr.HandleReconcile(container)
		}
	}
}

func (mgr *Manager) EligibleForReconcile(container *container.Container) bool {
	if container == nil {
		return false
	}

	c := container.Get()

	// Do work only on smr managed container others are left in piece to do they work
	if c.Labels["managed"] == "smr" {
		if container.Status.DefinitionDrift && !container.Status.Reconciling {
			return true
		}
	}

	return false
}

func (mgr *Manager) SmrManaged(container *container.Container) bool {
	if container == nil {
		return false
	}

	c := container.Get()

	if c.Labels["managed"] == "smr" {
		return true
	}

	return false
}

func (mgr *Manager) HandleReconcile(container *container.Container) {
	mgr.Registry.BackOffTracking(container.Static.Group, container.Static.Name)

	for {
		logger.Log.Info(fmt.Sprintf("trying to reconcile %s to the defined state", container.Static.GeneratedName))

		if mgr.Registry.BackOffTracker[container.Static.Group][container.Static.GeneratedName] > 5 {
			logger.Log.Error(fmt.Sprintf("%s container is backoff restarting", container.Static.Name))
			mgr.Registry.BackOffReset(container.Static.Group, container.Static.Name)
			container.Status.Reconciling = false
			return
		}

		if c := container.Get(); c != nil && c.State == "running" {
			container.Stop()

			timeout := false
			waitForStop := make(chan string, 1)
			go func() {
				for {
					c = container.Get()

					if timeout {
						return
					}

					if c != nil && c.State != "exited" {
						logger.Log.Info(fmt.Sprintf("waiting for container to exit %s", container.Static.GeneratedName))
						time.Sleep(1 * time.Second)
					} else {
						break
					}
				}

				waitForStop <- "container exited proceed with delete for reconciliation"
			}()

			select {
			case res := <-waitForStop:
				logger.Log.Info(fmt.Sprintf("%s %s", res, container.Static.GeneratedName))
			case <-time.After(30 * time.Second):
				logger.Log.Info("timed out waiting for the container to exit", zap.String("container", container.Static.GeneratedName))
				timeout = true
			}
		}

		err := container.Delete()

		if err == nil {
			var err error

			logger.Log.Info(fmt.Sprintf("trying to reconcile %s", container.Static.GeneratedName))

			container.Prepare(mgr.Badger)
			_, err = container.Run(mgr.Runtime, mgr.Badger, mgr.DnsCache)

			if err != nil {
				logger.Log.Error(fmt.Sprintf("Failed to reconcile %s to the defined state", container.Static.GeneratedName))
			} else {
				logger.Log.Info(fmt.Sprintf("reconcile of %s succeed", container.Static.GeneratedName))
				container.Status.Reconciling = false
				return
			}
		} else {
			logger.Log.Info(fmt.Sprintf("failed to delete %s container so that reconciliation can begin", container.Static.GeneratedName))
			logger.Log.Error(err.Error())
		}

		waitingTime := time.Duration(mgr.Registry.BackOffTracker[container.Static.Group][container.Static.Name]) * time.Second
		time.Sleep(waitingTime)
	}
}

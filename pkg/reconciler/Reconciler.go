package reconciler

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
	"smr/pkg/container"
	"smr/pkg/logger"
	"smr/pkg/registry"
	"smr/pkg/runtime"
	"smr/pkg/utils"
	"time"
)

func New() *Reconciler {
	return &Reconciler{
		QueueChan: make(chan Reconcile),
	}
}

func (reconciler *Reconciler) ListenEvents(registry *registry.Registry) {
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
			reconciler.Event(registry, msg)
		}
	}
}

func (reconciler *Reconciler) Event(registry *registry.Registry, event events.Message) {
	if utils.Contains([]string{"start", "kill", "stop", "die"}, event.Status) {
		container := registry.Find(event.Actor.Attributes["group"], event.Actor.Attributes["name"])

		c := container.Get()
		managed := false

		if c.Labels["managed"] == "smr" {
			managed = true
		}

		switch event.Status {
		case "start":
			if managed {
				reconciler.HandleStart(registry, container)
			}
			break
		case "kill":
			if managed {
				reconciler.HandleKill(registry, container)
			}
			break
		case "stop":
			if managed {
				reconciler.HandleStop(registry, container)
			}
			break
		case "die":
			if managed {
				reconciler.HandleDie(registry, container)
			}
			break
		default:
		}
	}
}

func (reconciler *Reconciler) HandleStart(registry *registry.Registry, container *container.Container) {
	// Container started it is running so update status accordingly
	container.Status.Reconciling = false
	container.Status.DefinitionDrift = false
	container.Status.Running = true
}

func (reconciler *Reconciler) HandleKill(registry *registry.Registry, container *container.Container) {
	// It can happen that kill signal occurs in the container even if it is not dying; eg killing thread, goroutine etc.
	container.Status.Running = true
}

func (reconciler *Reconciler) HandleStop(registry *registry.Registry, container *container.Container) {
	// Stop will stop the container so update the status accordingly
	container.Status.Running = false
}

func (reconciler *Reconciler) HandleDie(registry *registry.Registry, container *container.Container) {
	container.Status.Running = false

	if !container.Status.Reconciling {
		logger.Log.Info(fmt.Sprintf("sending event to queue for solving for container %s", container.Static.GeneratedName))
		reconciler.QueueChan <- Reconcile{
			Container: container,
		}
	}
}

func (reconciler *Reconciler) ListenQueue(registry *registry.Registry, runtime *runtime.Runtime, db *badger.DB, dnsCache map[string]string) {
	for {
		select {
		case queue := <-reconciler.QueueChan:
			logger.Log.Info(fmt.Sprintf("detected the event for reconciling %s", queue.Container.Static.GeneratedName))
			queue.Container.Status.Reconciling = true

			container := queue.Container
			registry.BackOffTracking(container.Static.Group, container.Static.Name)

			for {
				if registry.BackOffTracker[container.Static.Group][container.Static.GeneratedName] > 5 {
					logger.Log.Error(fmt.Sprintf("%s container is backoff restarting", container.Static.Name))

					registry.BackOffReset(container.Static.Group, container.Static.Name)
					container.Status.BackOffRestart = true
					container.Status.Healthy = false

					break
				}

				container.Stop()

				timeout := false
				waitForStop := make(chan string, 1)
				go func() {
					for {
						c := container.Get()

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

				err := container.Delete()

				if err == nil {
					container.Prepare(db)
					_, err = container.Run(runtime, db, dnsCache)

					break
				} else {

				}
			}

			queue.Container.Status.Reconciling = false
			break
		}
	}
}

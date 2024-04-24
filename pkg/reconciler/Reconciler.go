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
	"smr/pkg/dns"
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

func (reconciler *Reconciler) ListenEvents(registry *registry.Registry, dnsCache *dns.Records) {
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
			reconciler.Event(registry, dnsCache, msg)
		}
	}
}

func (reconciler *Reconciler) Event(registry *registry.Registry, dnsCache *dns.Records, event events.Message) {
	var container *container.Container

	// handle container events
	if utils.Contains([]string{"start", "kill", "stop", "die"}, event.Action) {
		container = registry.Find(event.Actor.Attributes["group"], event.Actor.Attributes["name"])
	}

	// handle network events
	if utils.Contains([]string{"connect", "disconnect"}, event.Action) {
		c := container.GetFromId(event.Actor.Attributes["container"])
		container = registry.Find(c.Labels["group"], c.Labels["name"])
	}

	if container == nil {
		return
	}

	c := container.Get()
	managed := false

	// only manage smr created containers, others are left alone to live and die in peace
	if c.Labels["managed"] == "smr" {
		managed = true
	}

	switch event.Action {
	case "connect":
		if managed {
			reconciler.HandleConnect(registry, dnsCache, container, event)
		}
		break
	case "disconnect":
		if managed {
			reconciler.HandleDisconnect(registry, dnsCache, container, event)
		}
		break
	case "start":
		if managed {
			reconciler.HandleStart(registry, container)
		}
		break
	case "kill":
		if managed {
			reconciler.HandleKill(registry, dnsCache, container)
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

func (reconciler *Reconciler) HandleConnect(registry *registry.Registry, dnsCache *dns.Records, container *container.Container, event events.Message) {
	// Handle network connect here
}

func (reconciler *Reconciler) HandleDisconnect(registry *registry.Registry, dnsCache *dns.Records, container *container.Container, event events.Message) {
	for _, ip := range dnsCache.FindDeleteQueue(container.GetDomain()) {
		dnsCache.RemoveARecord(container.GetDomain(), ip)
		dnsCache.RemoveARecord(container.GetHeadlessDomain(), ip)
	}

	dnsCache.ResetDeleteQueue(container.GetDomain())
}

func (reconciler *Reconciler) HandleStart(registry *registry.Registry, container *container.Container) {
	// Container started it is running so update status accordingly
	container.Status.Reconciling = false
	container.Status.DefinitionDrift = false
	container.Status.Running = true
}

func (reconciler *Reconciler) HandleKill(registry *registry.Registry, dnsCache *dns.Records, container *container.Container) {
	// It can happen that kill signal occurs in the container even if it is not dying; eg killing thread, goroutine etc.
	container.Status.Running = true

	for _, n := range container.Runtime.Networks {
		dnsCache.RemoveARecordQueue(container.GetDomain(), n.IP)
	}
}

func (reconciler *Reconciler) HandleStop(registry *registry.Registry, container *container.Container) {
	// Stop will stop the container so update the status accordingly
	container.Status.Running = false
}

func (reconciler *Reconciler) HandleDie(registry *registry.Registry, container *container.Container) {
	container.Status.Running = false

	reconcile := true

	// labels for ignoring events for specific container
	val, exists := container.Static.Labels["reconcile"]
	if exists {
		if val == "false" {
			logger.Log.Info("reconcile label set to false for the container, skipping reconcile", zap.String("container", container.Static.GeneratedName))
			reconcile = false
		}
	}

	if !container.Status.Reconciling && reconcile {
		logger.Log.Info(fmt.Sprintf("sending event to queue for solving for container %s", container.Static.GeneratedName))
		reconciler.QueueChan <- Reconcile{
			Container: container,
		}
	}
}

func (reconciler *Reconciler) ListenQueue(registry *registry.Registry, runtime *runtime.Runtime, db *badger.DB, dnsCache *dns.Records) {
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

func (reconciler *Reconciler) SmrManaged(container *container.Container) bool {
	if container == nil {
		return false
	}

	c := container.Get()

	if c.Labels["managed"] == "smr" {
		return true
	}

	return false
}

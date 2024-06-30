package reconciler

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/qdnqn/smr/pkg/container"
	"github.com/qdnqn/smr/pkg/dns"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/qdnqn/smr/pkg/registry"
	"github.com/qdnqn/smr/pkg/runtime"
	"github.com/qdnqn/smr/pkg/status"
	"github.com/qdnqn/smr/pkg/utils"
	"go.uber.org/zap"
	"time"
)

func New() *Reconciler {
	return &Reconciler{
		QueueChan:   make(chan Reconcile),
		QueueEvents: make(chan Events),
	}
}

func (reconciler *Reconciler) Event(registry *registry.Registry, dnsCache *dns.Records, event Events) {
	var container *container.Container

	// handle container events
	if utils.Contains([]string{"change"}, event.Kind) {
		container = registry.Find(event.Container.Static.Group, event.Container.Static.GeneratedName)
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

	switch event.Kind {
	case "change":
		if managed {
			reconciler.HandleChange(registry, dnsCache, container)
		}
		break
	default:
	}
}

func (reconciler *Reconciler) HandleChange(registry *registry.Registry, dnsCache *dns.Records, container *container.Container) {
	reconcile := true

	// labels for ignoring events for specific container
	val, exists := container.Static.Labels["reconcile"]
	if exists {
		if val == "false" {
			logger.Log.Info("reconcile label set to false for the container, skipping reconcile", zap.String("container", container.Static.GeneratedName))
			reconcile = false
		}
	}

	if !container.Status.TransitionState(status.STATUS_RECONCILING) && reconcile {
		logger.Log.Info(fmt.Sprintf("sending event to queue for solving for container %s", container.Static.GeneratedName))
		reconciler.QueueChan <- Reconcile{
			Container: container,
		}
	}
}

func (reconciler *Reconciler) ListenQueue(registry *registry.Registry, runtime *runtime.Runtime, db *badger.DB, dbEncrypted *badger.DB, dnsCache *dns.Records) {
	for {
		select {
		case queue := <-reconciler.QueueChan:
			logger.Log.Info(fmt.Sprintf("detected the event for reconciling %s", queue.Container.Static.GeneratedName))
			queue.Container.Status.TransitionState(status.STATUS_RECONCILING)

			container := queue.Container
			registry.BackOffTracking(container.Static.Group, container.Static.GeneratedName)

			for {
				if registry.BackOffTracker[container.Static.Group][container.Static.GeneratedName] > 5 {
					logger.Log.Error(fmt.Sprintf("%s container is backoff restarting", container.Static.GeneratedName))

					registry.BackOffReset(container.Static.Group, container.Static.GeneratedName)

					container.Status.TransitionState(status.STATUS_BACKOFF)

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

				if container.Status.IfStateIs(status.STATUS_BACKOFF) {
					logger.Log.Info("container is backoff restarting", zap.String("container", container.Static.GeneratedName))
				} else {
					if err == nil {
						if !container.Status.IfStateIs(status.STATUS_PENDING_DELETE) {
							container.Prepare(db)
							_, err = container.Run(runtime, db, dbEncrypted, dnsCache)
						} else {
							logger.Log.Info("container stopped and deleted", zap.String("container", container.Static.GeneratedName))
						}
						break
					} else {
						logger.Log.Info("failed to delete container", zap.String("container", container.Static.GeneratedName))
					}
				}
			}

			break
		}
	}
}

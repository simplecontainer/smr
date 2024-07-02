package reconcile

import (
	"context"
	"fmt"
	"github.com/qdnqn/smr/implementations/container/container"
	"github.com/qdnqn/smr/implementations/containers/replicas"
	"github.com/qdnqn/smr/implementations/containers/shared"
	"github.com/qdnqn/smr/implementations/containers/watcher"
	"github.com/qdnqn/smr/pkg/database"
	v1 "github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/qdnqn/smr/pkg/objects"
	"github.com/qdnqn/smr/pkg/registry"
	"github.com/r3labs/diff/v3"
	"time"
)

func NewWatcher(containers v1.Containers) *watcher.Containers {
	interval := 5 * time.Second
	ctx, fn := context.WithCancel(context.Background())

	return &watcher.Containers{
		Definition:      containers,
		Syncing:         false,
		Tracking:        false,
		ContainersQueue: make(chan string),
		Ctx:             ctx,
		Cancel:          fn,
		Ticker:          time.NewTicker(interval),
	}
}

func HandleTickerAndEvents(shared *shared.Shared, containers *watcher.Containers) {
	for {
		select {
		case <-containers.Ctx.Done():
			containers.Ticker.Stop()
			close(containers.ContainersQueue)
			shared.Watcher.Remove(fmt.Sprintf("%s.%s", containers.Definition.Meta.Group, containers.Definition.Meta.Name))

			return
		case _ = <-containers.ContainersQueue:
			ReconcileContainer(shared, containers)
			break
		case _ = <-containers.Ticker.C:
			if !containers.Syncing {
				ReconcileContainer(shared, containers)
			}
			break
		}
	}
}

func ReconcileContainer(shared *shared.Shared, containers *watcher.Containers) {
	if containers.Syncing {
		logger.Log.Info("containers already reconciling, waiting for the free slot")
		return
	}

	containers.Syncing = true
	var err error

	for _, definition := range containers.Definition.Spec {
		format := database.Format("container", definition.Meta.Group, definition.Meta.Name, "object")
		obj := objects.New()
		err = obj.Find(shared.Manager.Badger, format)

		if err != nil {
		}

		var jsonStringFromRequest string
		jsonStringFromRequest, err = definition.ToJsonString()

		if obj.Exists() {
			if obj.Diff(jsonStringFromRequest) {
				// Detect only change on replicas, if that's true tackle only scale up or scale down without recreating
				// containers that are there already, otherwise recreate everything
			}
		} else {
			err = obj.Add(shared.Manager.Badger, format, jsonStringFromRequest)
		}

		if obj.ChangeDetected() {
			err = obj.Update(shared.Manager.Badger, format, jsonStringFromRequest)
		}

		name := definition.Meta.Name

		_, ok := containers.Definition.Spec[name]

		if !ok {
			logger.Log.Error(fmt.Sprintf("container definintion with name %s not found", name))
			containers.Syncing = false
			return
		}

		groups, _, err := generateReplicaNamesAndGroups(shared, obj.ChangeDetected(), containers.Definition.Spec[name], obj.Changelog)

		if err == nil {
			if len(groups) > 0 {
				/*
					containerObjs := FetchContainersFromRegistry(shared.Registry, groups, names)

					for k, containerObj := range containerObjs {
						if obj.ChangeDetected() || !obj.Exists() {
							containerFromDefinition := reconcileContainer.NewWatcher(containerObjs[k])
							GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)

							ContainerTracker := shared.Watcher.Find(GroupIdentifier)

							containerFromDefinition.Container.Status.SetState(status.STATUS_CREATED)
							shared.Watcher.AddOrUpdate(GroupIdentifier, containerFromDefinition)

							if ContainerTracker == nil {
								go reconcileContainer.HandleTickerAndEvents(shared, containerFromDefinition)
							} else {
								if ContainerTracker.Tracking == false {
									ContainerTracker.Tracking = true
									go reconcileContainer.HandleTickerAndEvents(shared, containerFromDefinition)
								}
							}

							reconcileContainer.ReconcileContainer(shared, containerFromDefinition)


						} else {
							logger.Log.Info("no change detected in the containers definition")
						}
					}
				*/
			}
		} else {
			logger.Log.Error(err.Error())
			containers.Syncing = false
			return
		}
	}

	containers.Syncing = true
}

func generateReplicaNamesAndGroups(shared *shared.Shared, changed bool, containerDefinition v1.Container, changelog diff.Changelog) ([]string, []string, error) {
	_, index := shared.Registry.Name(containerDefinition.Meta.Group, containerDefinition.Meta.Name, shared.Manager.Runtime.PROJECT)

	r := replicas.Replicas{
		Group:          containerDefinition.Meta.Group,
		GeneratedIndex: index - 1,
		Replicas:       containerDefinition.Spec.Container.Replicas,
		Changed:        changed,
	}

	groups, names, err := r.HandleReplica(shared, containerDefinition, changelog)

	return groups, names, err
}

func GetReplicaNamesAndGroups(shared *shared.Shared, containerDefinition v1.Container) ([]string, []string, error) {
	_, index := shared.Registry.Name(containerDefinition.Meta.Group, containerDefinition.Meta.Name, shared.Manager.Runtime.PROJECT)

	r := replicas.Replicas{
		Group:          containerDefinition.Meta.Group,
		GeneratedIndex: index - 1,
		Replicas:       containerDefinition.Spec.Container.Replicas,
	}

	groups, names, err := r.GetReplica(shared, containerDefinition)

	return groups, names, err
}

func FetchContainersFromRegistry(registry *registry.Registry, groups []string, names []string) []*container.Container {
	var order []*container.Container

	for i, _ := range names {
		order = append(order, registry.Containers[groups[i]][names[i]])
	}

	return order
}

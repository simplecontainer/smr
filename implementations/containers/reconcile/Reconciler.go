package reconcile

import (
	"context"
	"fmt"
	reconcileContainer "github.com/qdnqn/smr/implementations/container/reconcile"
	"github.com/qdnqn/smr/implementations/containers/replicas"
	"github.com/qdnqn/smr/pkg/container"
	"github.com/qdnqn/smr/pkg/database"
	v1 "github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/objects"
	"github.com/qdnqn/smr/pkg/reconciler"
	"github.com/qdnqn/smr/pkg/registry"
	"github.com/qdnqn/smr/pkg/status"
	"github.com/r3labs/diff/v3"
	"time"
)

func NewWatcher(containers v1.Containers) *reconciler.Containers {
	interval := 5 * time.Second
	ctx, fn := context.WithCancel(context.Background())

	return &reconciler.Containers{
		Definition:     containers,
		Syncing:        false,
		Tracking:       false,
		ContainerQueue: make(chan reconciler.Event),
		Ctx:            ctx,
		Cancel:         fn,
		Ticker:         time.NewTicker(interval),
	}
}

func HandleTickerAndEvents(mgr *manager.Manager, containers *reconciler.Containers) {
	for {
		select {
		case <-containers.Ctx.Done():
			containers.Ticker.Stop()
			close(containers.ContainerQueue)
			mgr.ContainersWatchers.Remove(fmt.Sprintf("%s.%s", containers.Definition.Meta.Group, containers.Definition.Meta.Name))

			return
		case _ = <-containers.ContainerQueue:
			ReconcileContainer(mgr, containers)
			break
		case _ = <-containers.Ticker.C:
			if !containers.Syncing {
				ReconcileContainer(mgr, containers)
			}
			break
		}
	}
}

func ReconcileContainer(mgr *manager.Manager, containers *reconciler.Containers) {
	if containers.Syncing {
		logger.Log.Info("containers already reconciling, waiting for the free slot")
		return
	}

	containers.Syncing = true
	var err error

	for _, definition := range containers.Definition.Spec {
		format := database.Format("container", definition.Meta.Group, definition.Meta.Name, "object")
		obj := objects.New()
		err = obj.Find(mgr.Badger, format)

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
			err = obj.Add(mgr.Badger, format, jsonStringFromRequest)
		}

		if obj.ChangeDetected() {
			err = obj.Update(mgr.Badger, format, jsonStringFromRequest)
		}

		name := definition.Meta.Name

		_, ok := containers.Definition.Spec[name]

		if !ok {
			logger.Log.Error(fmt.Sprintf("container definintion with name %s not found", name))
			containers.Syncing = false
			return
		}

		groups, names, err := generateReplicaNamesAndGroups(mgr, obj.ChangeDetected(), containers.Definition.Spec[name], obj.Changelog)

		if err == nil {
			if len(groups) > 0 {
				containerObjs := FetchContainersFromRegistry(mgr.Registry, groups, names)

				for k, containerObj := range containerObjs {
					if obj.ChangeDetected() || !obj.Exists() {
						containerFromDefinition := reconcileContainer.NewWatcher(containerObjs[k])
						GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)

						ContainerTracker := mgr.ContainerWatchers.Find(GroupIdentifier)

						containerFromDefinition.Container.Status.SetState(status.STATUS_CREATED)
						mgr.ContainerWatchers.AddOrUpdate(GroupIdentifier, containerFromDefinition)

						if ContainerTracker == nil {
							go reconcileContainer.HandleTickerAndEvents(mgr, containerFromDefinition)
						} else {
							if ContainerTracker.Tracking == false {
								ContainerTracker.Tracking = true
								go reconcileContainer.HandleTickerAndEvents(mgr, containerFromDefinition)
							}
						}

						reconcileContainer.ReconcileContainer(mgr, containerFromDefinition)
					} else {
						logger.Log.Info("no change detected in the containers definition")
					}
				}
			}
		} else {
			logger.Log.Error(err.Error())
			containers.Syncing = false
			return
		}
	}

	containers.Syncing = true
}

func generateReplicaNamesAndGroups(mgr *manager.Manager, changed bool, containerDefinition v1.Container, changelog diff.Changelog) ([]string, []string, error) {
	_, index := mgr.Registry.Name(containerDefinition.Meta.Group, containerDefinition.Meta.Name, mgr.Runtime.PROJECT)

	r := replicas.Replicas{
		Group:          containerDefinition.Meta.Group,
		GeneratedIndex: index - 1,
		Replicas:       containerDefinition.Spec.Container.Replicas,
		Changed:        changed,
	}

	groups, names, err := r.HandleReplica(mgr, containerDefinition, changelog)

	return groups, names, err
}

func GetReplicaNamesAndGroups(mgr *manager.Manager, containerDefinition v1.Container, changelog diff.Changelog) ([]string, []string, error) {
	_, index := mgr.Registry.Name(containerDefinition.Meta.Group, containerDefinition.Meta.Name, mgr.Runtime.PROJECT)

	r := replicas.Replicas{
		Group:          containerDefinition.Meta.Group,
		GeneratedIndex: index - 1,
		Replicas:       containerDefinition.Spec.Container.Replicas,
	}

	groups, names, err := r.GetReplica(mgr, containerDefinition, changelog)

	return groups, names, err
}

func FetchContainersFromRegistry(registry *registry.Registry, groups []string, names []string) []*container.Container {
	var order []*container.Container

	for i, _ := range names {
		order = append(order, registry.Containers[groups[i]][names[i]])
	}

	return order
}

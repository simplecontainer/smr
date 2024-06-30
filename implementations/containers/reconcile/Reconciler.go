package reconcile

import (
	"context"
	"fmt"
	"github.com/qdnqn/smr/pkg/database"
	v1 "github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/dependency"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/qdnqn/smr/pkg/objects"
	"github.com/qdnqn/smr/pkg/reconciler"
	"github.com/qdnqn/smr/pkg/replicas"
	"github.com/qdnqn/smr/pkg/status"
	"github.com/r3labs/diff/v3"
	"go.uber.org/zap"
	"time"
)

func NewWatcher(containers *v1.Containers) *reconciler.Containers {
	interval, err := time.ParseDuration("5s")

	if err != nil {
		logger.Log.Error(err.Error())
		return nil
	}

	return &reconciler.Containers{
		Definition:     containers,
		InSync:         false,
		ContainerQueue: make(chan reconciler.Event),
		Ctx:            context.Background(),
		Ticker:         time.NewTicker(interval),
	}
}

func HandleTickerAndEvents(mgr *manager.Manager, containers *reconciler.Containers) {
	for {
		select {
		case <-containers.Ctx.Done():
			return
			break
		case _ = <-containers.ContainerQueue:
			break
		case _ = <-containers.Ticker.C:
			ReconcileContainer(mgr, containers)
			break
		}
	}
}

func ReconcileContainer(mgr *manager.Manager, containers *reconciler.Containers) {
	var globalGroups []string
	var globalNames []string
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

		logger.Log.Info("object is changed", zap.String("container", definition.Meta.Name))

		name := definition.Meta.Name
		logger.Log.Info(fmt.Sprintf("trying to generate container %s object", name))

		_, ok := containers.Definition.Spec[name]

		if !ok {
			logger.Log.Error(fmt.Sprintf("container definintion with name %s not found", name))
			return
		}

		groups, names, err := generateReplicaNamesAndGroups(mgr, containers.Definition.Spec[name], obj.Changelog)

		if err == nil {
			logger.Log.Info(fmt.Sprintf("generated container %s object", name))

			globalGroups = append(globalGroups, groups...)
			globalNames = append(globalNames, names...)

			err = obj.Update(mgr.Badger, format, jsonStringFromRequest)
		} else {
			logger.Log.Error("failed to generate names and groups")
			return
		}
	}

	if len(globalGroups) > 0 {
		logger.Log.Info(fmt.Sprintf("trying to order containers by dependencies"))
		order := orderByDependencies(mgr.Registry, globalGroups, globalNames)
		logger.Log.Info(fmt.Sprintf("containers are ordered by dependencies"))

		for _, container := range order {

			switch container.Status.GetState() {
			case status.STATUS_CREATED:
				container.Status.TransitionState(status.STATUS_DEPENDS_SOLVING)
				solved, _ := dependency.Ready(mgr, container.Static.Group, container.Static.GeneratedName, container.Static.Definition.Spec.Container.Dependencies)

				if solved {
					logger.Log.Info("trying to run container", zap.String("group", container.Static.Group), zap.String("name", container.Static.Name))

					// Fix GitOps reconcile!!!!!!!!
					//container.SetOwner(c.Request.Header.Get("Owner"))
					container.Prepare(mgr.Badger)
					_, err := container.Run(mgr.Runtime, mgr.Badger, mgr.BadgerEncrypted, mgr.DnsCache)

					if err != nil {
						format := database.Format("container", container.Static.Group, container.Static.Name, "object")

						// clear the object in the store since container failed to run
						obj := objects.New()
						obj.Update(mgr.Badger, format, "")

						mgr.Registry.Remove(container.Static.Group, container.Static.GeneratedName)

						return
					}

					client, err := mgr.Keys.GenerateHttpClient()
					container.Ready(mgr.BadgerEncrypted, client, err)
				} else {
					logger.Log.Info("failed to solve container")
				}
				break
			case status.STATUS_PENDING_DELETE:
				logger.Log.Info(fmt.Sprintf("container is pending to delete %s", container.Static.GeneratedName))

				mgr.Registry.Remove(container.Static.Group, container.Static.GeneratedName)

				mgr.Reconciler.QueueChan <- reconciler.Reconcile{
					Container: container,
				}
				break
			case status.STATUS_DRIFTED:
				logger.Log.Info("sending container to reconcile state", zap.String("container", container.Static.GeneratedName))

				mgr.Reconciler.QueueChan <- reconciler.Reconcile{
					Container: container,
				}
				break
			}
		}
	}
}

func generateReplicaNamesAndGroups(mgr *manager.Manager, containerDefinition v1.Container, changelog diff.Changelog) ([]string, []string, error) {
	_, index := mgr.Registry.Name(containerDefinition.Meta.Group, containerDefinition.Meta.Name, mgr.Runtime.PROJECT)

	r := replicas.Replicas{
		Group:          containerDefinition.Meta.Group,
		GeneratedIndex: index - 1,
		Replicas:       containerDefinition.Spec.Container.Replicas,
	}

	groups, names, err := r.HandleReplica(mgr, containerDefinition, changelog)

	return groups, names, err
}

func getReplicaNamesAndGroups(mgr *manager.Manager, containerDefinition v1.Container, changelog diff.Changelog) ([]string, []string, error) {
	_, index := mgr.Registry.Name(containerDefinition.Meta.Group, containerDefinition.Meta.Name, mgr.Runtime.PROJECT)

	r := replicas.Replicas{
		Group:          containerDefinition.Meta.Group,
		GeneratedIndex: index - 1,
		Replicas:       containerDefinition.Spec.Container.Replicas,
	}

	groups, names, err := r.GetReplica(mgr, containerDefinition, changelog)

	return groups, names, err
}

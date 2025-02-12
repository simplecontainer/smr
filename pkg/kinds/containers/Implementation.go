package containers

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/events/platform"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/replicas"
	"github.com/simplecontainer/smr/pkg/kinds/containers/reconcile"
	"github.com/simplecontainer/smr/pkg/kinds/containers/registry"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/wI2L/jsondiff"
	"net/http"
	"os"
)

func (containers *Containers) Start() error {
	containers.Started = true

	containers.Shared.Watchers = &watcher.Containers{}
	containers.Shared.Watchers.Watchers = make(map[string]*watcher.Container)

	containers.Shared.Registry = registry.New(containers.Shared.Client, containers.Shared.User)

	logger.Log.Info(fmt.Sprintf("platform for running containers is %s", containers.Shared.Manager.Config.Platform))

	// Check if everything alright with the daemon
	switch containers.Shared.Manager.Config.Platform {
	case static.PLATFORM_DOCKER:
		if err := docker.IsDaemonRunning(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		break
	}

	// Start listening events based on the platform and for internal events
	go platform.Listen(containers.Shared, containers.Shared.Manager.Config.Platform)

	logger.Log.Info(fmt.Sprintf("started listening events for simplecontainer and platform: %s", containers.Shared.Manager.Config.Platform))

	return nil
}
func (containers *Containers) GetShared() interface{} {
	return containers.Shared
}

func (containers *Containers) Apply(user *authentication.User, definition []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CONTAINERS, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	obj, err := request.Apply(containers.Shared.Client, user)

	if !obj.ChangeDetected() {
		return common.Response(http.StatusOK, static.STATUS_RESPONSE_APPLIED, nil, nil), nil
	}

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	var create []platforms.IContainer
	var update []platforms.IContainer
	var destroy []platforms.IContainer

	create, update, destroy, err = GenerateContainers(containers.Shared, request.Definition.Definition.(*v1.ContainersDefinition), obj.GetDiff())

	if err != nil {
		return common.Response(http.StatusInternalServerError, "failed to generate replica counts", err, nil), err
	}

	if len(destroy) > 0 {
		for _, containerObj := range destroy {
			GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.GetGroup(), containerObj.GetGeneratedName())

			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PENDING_DELETE)
			reconcile.Containers(containers.Shared, containers.Shared.Watchers.Find(GroupIdentifier))
		}
	}

	if len(update) > 0 {
		if len(obj.GetDiff()) == 1 {
			if obj.GetDiff()[0].Path == "/spec/replicas" {
				update = nil
			}
		}

		for _, containerObj := range update {
			if obj.Exists() {
				existingWatcher := containers.Shared.Watchers.Find(containerObj.GetGroupIdentifier())

				if existingWatcher != nil {
					existingWatcher.Logger.Info("container object modified, reusing watcher")
					existingContainer := containers.Shared.Registry.Find(containerObj.GetDefinition().GetPrefix(), containerObj.GetGroup(), containerObj.GetGeneratedName())

					if existingContainer != nil {
						existingWatcher.Ticker.Stop()

						containerObj.GetStatus().SetState(status.STATUS_CREATED)
						containers.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)

						existingWatcher.Container = containerObj
						go reconcile.Containers(containers.Shared, existingWatcher)
					}
				} else {
					existingWatcher.Logger.Info("no changes detected on the container object")
				}
			} else {
				// forbiden update occured
			}
		}
	}

	if len(create) > 0 {
		for _, containerObj := range create {
			if obj.Exists() {
				existingWatcher := containers.Shared.Watchers.Find(containerObj.GetGroupIdentifier())

				if existingWatcher != nil {
					// forbiden create occured
				} else {
					// container object holding multiple replicas existed but watcher was never born
					// this means that replica landed on this node first time, could be that it comes from 2nd node
					// implication:
					// - create watcher
					// - assing container to the wathcer
					// - roll the reconciler

					existingContainer := containers.Shared.Registry.Find(containerObj.GetDefinition().GetPrefix(), containerObj.GetGroup(), containerObj.GetGeneratedName())

					if existingContainer != nil && existingContainer.IsGhost() {
						w := watcher.New(containerObj, status.STATUS_TRANSFERING, user)
						containers.Shared.Watchers.AddOrUpdate(containerObj.GetGroupIdentifier(), w)
						containers.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)

						w.Logger.Info("container object created")

						go reconcile.HandleTickerAndEvents(containers.Shared, w, func(w *watcher.Container) error {
							return nil
						})

						go reconcile.Containers(containers.Shared, w)
					} else {
						w := watcher.New(containerObj, status.STATUS_CREATED, user)
						containers.Shared.Watchers.AddOrUpdate(containerObj.GetGroupIdentifier(), w)
						containers.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)

						w.Logger.Info("container object created")

						go reconcile.HandleTickerAndEvents(containers.Shared, w, func(w *watcher.Container) error {
							return nil
						})
						go reconcile.Containers(containers.Shared, w)
					}
				}
			} else {
				// container object holding multiple replicas never existed
				// implication:
				// - create watcher
				// - assing container to the wathcer
				// - roll the reconciler

				w := watcher.New(containerObj, status.STATUS_CREATED, user)
				containers.Shared.Watchers.AddOrUpdate(containerObj.GetGroupIdentifier(), w)
				containers.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)

				w.Logger.Info("container object created")

				go reconcile.HandleTickerAndEvents(containers.Shared, w, func(w *watcher.Container) error {
					return nil
				})

				go reconcile.Containers(containers.Shared, w)
			}
		}
	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}
func (containers *Containers) Compare(user *authentication.User, definition []byte) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CONTAINERS, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Apply(containers.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusTeapot, "object drifted", nil, nil), nil
	} else {
		return common.Response(http.StatusOK, "object in sync", nil, nil), nil
	}
}
func (containers *Containers) Delete(user *authentication.User, definition []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CONTAINERS, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Remove(containers.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusTeapot, "", err, nil), err
	}

	var destroy []platforms.IContainer
	destroy, err = GetContainers(containers.Shared, request.Definition.Definition.(*v1.ContainersDefinition))

	if err != nil {
		return common.Response(http.StatusInternalServerError, "failed to generate replica counts", err, nil), err
	}

	if err == nil {
		if len(destroy) > 0 {
			for _, containerObj := range destroy {
				go func() {
					GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.GetGroup(), containerObj.GetGeneratedName())
					containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PENDING_DELETE)

					reconcile.Containers(containers.Shared, containers.Shared.Watchers.Find(GroupIdentifier))
				}()
			}

			return common.Response(http.StatusOK, "object is deleted", nil, nil), nil
		} else {
			return common.Response(http.StatusNotFound, "object not found", errors.New("object not found"), nil), errors.New("object not found")
		}
	} else {
		return common.Response(http.StatusNotFound, "object not found", errors.New("object not found"), nil), errors.New("object not found")
	}
}
func (containers *Containers) Event(event contracts.Event) error {
	switch event.GetType() {
	case events.EVENT_CHANGE:
		for _, containerWatcher := range containers.Shared.Watchers.Watchers {
			if containerWatcher.Container.HasDependencyOn(event.GetKind(), event.GetGroup(), event.GetName()) {
				containerWatcher.Container.GetStatus().TransitionState(containerWatcher.Container.GetGroup(), containerWatcher.Container.GetGeneratedName(), status.STATUS_CHANGE)
				containers.Shared.Watchers.Find(containerWatcher.Container.GetGroupIdentifier()).ContainerQueue <- containerWatcher.Container
			}
		}
		break
	case events.EVENT_RESTART:
		containerObj := containers.Shared.Registry.FindLocal(event.GetGroup(), event.GetName())

		if containerObj == nil {
			return errors.New("container not found event is ignored")
		}

		containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_CREATED)
		containers.Shared.Watchers.Find(containerObj.GetGroupIdentifier()).ContainerQueue <- containerObj

		break
	}

	return nil
}

func GenerateContainers(shared *shared.Shared, definition *v1.ContainersDefinition, changelog jsondiff.Patch) ([]platforms.IContainer, []platforms.IContainer, []platforms.IContainer, error) {
	r := replicas.New(shared.Manager.Cluster.Node.NodeID, shared.Manager.Cluster.Cluster.Nodes)
	return r.GenerateContainers(shared.Registry, definition, shared.Manager.Config)
}

func GetContainers(shared *shared.Shared, definition *v1.ContainersDefinition) ([]platforms.IContainer, error) {
	r := replicas.New(shared.Manager.Cluster.Node.NodeID, shared.Manager.Cluster.Cluster.Nodes)
	return r.RemoveContainers(shared.Registry, definition)
}

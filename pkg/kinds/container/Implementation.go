package container

import (
	"errors"
	"fmt"
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/events/platform"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/replicas"
	"github.com/simplecontainer/smr/pkg/kinds/container/reconcile"
	"github.com/simplecontainer/smr/pkg/kinds/container/registry"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/kinds/container/status"
	"github.com/simplecontainer/smr/pkg/kinds/container/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
	"os"
)

func (container *Container) Start() error {
	container.Started = true

	container.Shared.Watchers = &watcher.Containers{}
	container.Shared.Watchers.Watchers = make(map[string]*watcher.Container)

	container.Shared.Registry = &registry.Registry{
		Containers:     make(map[string]map[string]platforms.IContainer),
		Indexes:        make(map[string][]uint64),
		BackOffTracker: make(map[string]map[string]uint64),
		Client:         container.Shared.Client,
		User:           container.Shared.User,
	}

	logger.Log.Info(fmt.Sprintf("platform for running container is %s", container.Shared.Manager.Config.Platform))

	// Check if everything alright with the daemon
	switch container.Shared.Manager.Config.Platform {
	case static.PLATFORM_DOCKER:
		if err := docker.IsDaemonRunning(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		break
	}

	// Start listening events based on the platform and for internal events
	go platform.Listen(container.Shared, container.Shared.Manager.Config.Platform)

	logger.Log.Info(fmt.Sprintf("started listening events for simplecontainer and platform: %s", container.Shared.Manager.Config.Platform))

	return nil
}
func (container *Container) GetShared() interface{} {
	return container.Shared
}

func (container *Container) Apply(user *authentication.User, definition []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CONTAINER, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	obj, err := request.Apply(container.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	var create []platforms.IContainer
	var update []platforms.IContainer
	var destroy []platforms.IContainer

	create, update, destroy, err = GenerateContainers(container.Shared, request.Definition.Definition.(*v1.ContainerDefinition), obj.GetDiff())

	if err != nil {
		return common.Response(http.StatusInternalServerError, "failed to generate replica counts", err, nil), err
	}

	if len(destroy) > 0 {
		for _, containerObj := range destroy {
			GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.GetGroup(), containerObj.GetGeneratedName())

			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PENDING_DELETE)
			reconcile.Container(container.Shared, container.Shared.Watchers.Find(GroupIdentifier))
		}
	}

	if len(update) > 0 {
		for _, containerObj := range create {
			if obj.Exists() {
				existingWatcher := container.Shared.Watchers.Find(containerObj.GetGroupIdentifier())

				if obj.ChangeDetected() {
					if existingWatcher != nil {
						existingWatcher.Logger.Info("container object modified, reusing watcher")
						existingContainer := container.Shared.Registry.Find(containerObj.GetDefinition().GetPrefix(), containerObj.GetGroup(), containerObj.GetGeneratedName())

						if existingContainer != nil {
							existingWatcher.Ticker.Stop()

							containerObj.GetStatus().Logger = existingWatcher.Logger
							containerObj.GetStatus().SetState(status.STATUS_CREATED)
							container.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)

							existingWatcher.Container = containerObj
							existingWatcher.ContainerQueue <- containerObj
						}
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
				existingWatcher := container.Shared.Watchers.Find(containerObj.GetGroupIdentifier())

				if existingWatcher != nil {
					// forbiden create occured
				} else {
					// container object holding multiple replicas existed but watcher was never born
					// this means that replica landed on this node first time, could be that it comes from 2nd node
					// implication:
					// - create watcher
					// - assing container to the wathcer
					// - roll the reconciler

					existingContainer := container.Shared.Registry.Find(containerObj.GetDefinition().GetPrefix(), containerObj.GetGroup(), containerObj.GetGeneratedName())

					if existingContainer != nil && existingContainer.IsGhost() {
						watcher := watcher.New(containerObj, status.STATUS_TRANSFERING, user)
						container.Shared.Watchers.AddOrUpdate(containerObj.GetGroupIdentifier(), watcher)
						container.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)

						watcher.Logger.Info("container object created")

						go reconcile.HandleTickerAndEvents(container.Shared, watcher)
						go reconcile.Container(container.Shared, watcher)
					} else {
						watcher := watcher.New(containerObj, status.STATUS_CREATED, user)
						container.Shared.Watchers.AddOrUpdate(containerObj.GetGroupIdentifier(), watcher)
						container.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)

						watcher.Logger.Info("container object created")

						go reconcile.HandleTickerAndEvents(container.Shared, watcher)
						go reconcile.Container(container.Shared, watcher)
					}
				}
			} else {
				// container object holding multiple replicas never existed
				// implication:
				// - create watcher
				// - assing container to the wathcer
				// - roll the reconciler

				watcher := watcher.New(containerObj, status.STATUS_CREATED, user)
				container.Shared.Watchers.AddOrUpdate(containerObj.GetGroupIdentifier(), watcher)
				container.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)

				watcher.Logger.Info("container object created")

				go reconcile.HandleTickerAndEvents(container.Shared, watcher)
				go reconcile.Container(container.Shared, watcher)
			}
		}
	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}
func (container *Container) Compare(user *authentication.User, definition []byte) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CONTAINER, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Apply(container.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusTeapot, "object drifted", nil, nil), nil
	} else {
		return common.Response(http.StatusOK, "object in sync", nil, nil), nil
	}
}
func (container *Container) Delete(user *authentication.User, definition []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CONTAINER, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Remove(container.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusTeapot, "", err, nil), err
	}

	var destroy []platforms.IContainer
	destroy, err = GetContainers(container.Shared, request.Definition.Definition.(*v1.ContainerDefinition))

	if err != nil {
		return common.Response(http.StatusInternalServerError, "failed to generate replica counts", err, nil), err
	}

	if err == nil {
		if len(destroy) > 0 {
			for _, containerObj := range destroy {
				go func() {
					GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.GetGroup(), containerObj.GetGeneratedName())
					containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PENDING_DELETE)

					reconcile.Container(container.Shared, container.Shared.Watchers.Find(GroupIdentifier))
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
func (container *Container) Event(event contracts.Event) error {
	switch event.GetType() {
	case events.EVENT_CHANGE:
		for _, containerWatcher := range container.Shared.Watchers.Watchers {
			if containerWatcher.Container.HasDependencyOn(event.GetKind(), event.GetGroup(), event.GetName()) {
				containerWatcher.Container.GetStatus().TransitionState(containerWatcher.Container.GetGroup(), containerWatcher.Container.GetGeneratedName(), status.STATUS_CHANGE)
				container.Shared.Watchers.Find(containerWatcher.Container.GetGroupIdentifier()).ContainerQueue <- containerWatcher.Container
			}
		}
		break
	case events.EVENT_RESTART:
		containerObj := container.Shared.Registry.FindLocal(event.GetGroup(), event.GetName())

		if containerObj == nil {
			return errors.New("container not found event is ignored")
		}

		containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_CREATED)
		container.Shared.Watchers.Find(containerObj.GetGroupIdentifier()).ContainerQueue <- containerObj

		break
	}

	return nil
}

func GenerateContainers(shared *shared.Shared, definition *v1.ContainerDefinition, changelog diff.Changelog) ([]platforms.IContainer, []platforms.IContainer, []platforms.IContainer, error) {
	r := replicas.New(shared.Manager.Cluster.Node.NodeID, shared.Manager.Cluster.Cluster.Nodes)
	return r.GenerateContainers(shared.Registry, definition, shared.Manager.Config)
}

func GetContainers(shared *shared.Shared, definition *v1.ContainerDefinition) ([]platforms.IContainer, error) {
	r := replicas.New(shared.Manager.Cluster.Node.NodeID, shared.Manager.Cluster.Cluster.Nodes)
	return r.RemoveContainers(shared.Registry, definition)
}

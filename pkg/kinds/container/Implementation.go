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
	"github.com/simplecontainer/smr/pkg/f"
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
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func (container *Container) Start() error {
	container.Started = true

	container.Shared.Watcher = &watcher.ContainerWatcher{}
	container.Shared.Watcher.Container = make(map[string]*watcher.Container)

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

func (container *Container) Apply(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_CONTAINER)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.ContainerDefinition)

	_, err = definition.Validate()

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	format := f.New(definition.GetPrefix(), static.CATEGORY_KIND, static.KIND_CONTAINER, definition.Meta.Group, definition.Meta.Name)
	obj := objects.New(container.Shared.Client.Get(user.Username), user)

	var jsonStringFromRequest []byte
	jsonStringFromRequest, err = definition.ToJson()

	logger.Log.Debug("server received container object", zap.String("definition", string(jsonStringFromRequest)))

	var create []platforms.IContainer
	var destroy []platforms.IContainer

	obj, err = request.Definition.Apply(format, obj, static.KIND_CONTAINER)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	} else {
		create, destroy, err = GenerateContainers(container.Shared, definition, obj.GetDiff())

		if err != nil {
			return common.Response(http.StatusInternalServerError, "failed to generate replica counts", err, nil), err
		}
	}

	if len(destroy) > 0 {
		for _, containerObj := range destroy {
			GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.GetGroup(), containerObj.GetGeneratedName())

			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PENDING_DELETE)
			reconcile.Container(container.Shared, container.Shared.Watcher.Find(GroupIdentifier))
		}
	}

	if len(create) > 0 {
		for _, containerObj := range create {
			GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.GetGroup(), containerObj.GetGeneratedName())
			existingWatcher := container.Shared.Watcher.Find(GroupIdentifier)

			if obj.Exists() {
				if obj.ChangeDetected() {
					existingWatcher = reconcile.NewWatcher(containerObj, container.Shared.Manager, user)
					existingWatcher.Logger.Info("container object modified")

					existingContainer := container.Shared.Registry.Find(containerObj.GetDefinition().GetPrefix(), containerObj.GetGroup(), containerObj.GetGeneratedName())
					if existingContainer != nil && existingContainer.IsGhost() {
						existingWatcher.Container.GetStatus().SetState(status.STATUS_TRANSFERING)
					} else {
						existingWatcher.Container.GetStatus().SetState(status.STATUS_CREATED)
						container.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)
					}

					go reconcile.HandleTickerAndEvents(container.Shared, existingWatcher)
					container.Shared.Watcher.AddOrUpdate(GroupIdentifier, existingWatcher)

					go reconcile.Container(container.Shared, existingWatcher)
				} else {
					logger.Log.Info("no change detected in the containers definition")
				}
			} else {
				_, err = containerObj.GetContainerState()
				container.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)

				if err != nil {
					existingWatcher = reconcile.NewWatcher(containerObj, container.Shared.Manager, user)
					existingWatcher.Logger.Info("container object created")

					go reconcile.HandleTickerAndEvents(container.Shared, existingWatcher)

					existingWatcher.Container.GetStatus().SetState(status.STATUS_CREATED)
					container.Shared.Watcher.AddOrUpdate(GroupIdentifier, existingWatcher)
				} else {
					existingWatcher = reconcile.NewWatcher(containerObj, container.Shared.Manager, user)
					existingWatcher.Logger.Info("container object created but already is existing - manual intervention needed")

					go reconcile.HandleTickerAndEvents(container.Shared, existingWatcher)

					existingWatcher.Container.GetStatus().SetState(status.STATUS_RECREATED)
					container.Shared.Watcher.AddOrUpdate(GroupIdentifier, existingWatcher)
				}

				go reconcile.Container(container.Shared, existingWatcher)
			}
		}
	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}
func (container *Container) Compare(user *authentication.User, jsonData []byte) (contracts.Response, error) {
	return common.Response(http.StatusOK, "object in sync", nil, nil), nil
}
func (container *Container) Delete(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_CONTAINER)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	definition := request.Definition.Definition.(*v1.ContainerDefinition)

	format := f.New(definition.GetPrefix(), static.CATEGORY_KIND, static.KIND_CONTAINER, definition.Meta.Group, definition.Meta.Name)
	obj := objects.New(container.Shared.Client.Get(user.Username), user)

	var destroy []platforms.IContainer

	existingDefinition, err := request.Definition.Delete(format, obj, static.KIND_CONTAINER)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	} else {
		destroy, err = GetContainers(container.Shared, existingDefinition.(*v1.ContainerDefinition))

		if err != nil {
			return common.Response(http.StatusInternalServerError, "failed to generate replica counts", err, nil), err
		}
	}

	if err == nil {
		if len(destroy) > 0 {
			for _, containerObj := range destroy {
				go func() {
					GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.GetGroup(), containerObj.GetGeneratedName())
					containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PENDING_DELETE)

					reconcile.Container(container.Shared, container.Shared.Watcher.Find(GroupIdentifier))
				}()
			}

			return common.Response(http.StatusOK, "object is deleted", nil, nil), nil
		} else {
			return common.Response(http.StatusNotFound, "container is not found on the server definition sent", errors.New("container not found"), nil), errors.New("container not found")
		}
	} else {
		return common.Response(http.StatusNotFound, "container is not found on the server definition sent", errors.New("container not found"), nil), errors.New("container not found")
	}
}
func (container *Container) Event(event contracts.Event) error {
	switch event.GetType() {
	case events.EVENT_DELETE:

		break
	case events.EVENT_CHANGE:
		for _, containerWatcher := range container.Shared.Watcher.Container {
			if containerWatcher.Container.HasDependencyOn(event.GetKind(), event.GetGroup(), event.GetName()) {
				containerWatcher.Container.GetStatus().TransitionState(containerWatcher.Container.GetGroup(), containerWatcher.Container.GetGeneratedName(), status.STATUS_CHANGE)
				container.Shared.Watcher.Find(containerWatcher.Container.GetGroupIdentifier()).ContainerQueue <- containerWatcher.Container
			}
		}
		break
	case events.EVENT_RESTART:
		containerObj := container.Shared.Registry.FindLocal(event.GetGroup(), event.GetName())

		if containerObj == nil {
			return errors.New("container not found event is ignored")
		}

		containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_CREATED)
		container.Shared.Watcher.Find(containerObj.GetGroupIdentifier()).ContainerQueue <- containerObj

		break
	}

	return nil
}

func GenerateContainers(shared *shared.Shared, definition *v1.ContainerDefinition, changelog diff.Changelog) ([]platforms.IContainer, []platforms.IContainer, error) {
	r := replicas.New(shared.Manager.Cluster.Node.NodeID, shared.Manager.Cluster.Cluster.Nodes)
	return r.GenerateContainers(shared.Registry, definition, shared.Manager.Config)
}

func GetContainers(shared *shared.Shared, definition *v1.ContainerDefinition) ([]platforms.IContainer, error) {
	r := replicas.New(shared.Manager.Cluster.Node.NodeID, shared.Manager.Cluster.Cluster.Nodes)
	return r.RemoveContainers(shared.Registry, definition)
}

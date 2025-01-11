package container

import (
	"errors"
	"fmt"
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	replication "github.com/simplecontainer/smr/pkg/distributed"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/container/distributed"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/events"
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
	"reflect"
	"strings"
)

func (container *Container) Start() error {
	container.Started = true

	container.Shared.Watcher = &watcher.ContainerWatcher{
		EventChannel: make(chan replication.KV),
	}
	container.Shared.Watcher.Container = make(map[string]*watcher.Container)

	container.Shared.Registry = &registry.Registry{
		ChangeC:        make(chan distributed.Container),
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
		docker.IsDaemonRunning()
		break
	}

	// Start listening events based on the platform and for internal events
	go events.NewPlatformEventsListener(container.Shared, container.Shared.Manager.Config.Platform)
	go events.NewEventsListener(container.Shared, container.Shared.Watcher.EventChannel)

	go container.Shared.Registry.ListenChanges()

	logger.Log.Info(fmt.Sprintf("started listening events for simplecontainer and platform: %s", container.Shared.Manager.Config.Platform))

	return nil
}
func (container *Container) GetShared() interface{} {
	return container.Shared
}
func (container *Container) Apply(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_CONTAINER)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err), err
	}

	definition := request.Definition.Definition.(*v1.ContainerDefinition)

	_, err = definition.Validate()

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err), err
	}

	format := f.New("container", definition.Meta.Group, definition.Meta.Name, "object")
	obj := objects.New(container.Shared.Client.Get(user.Username), user)

	var jsonStringFromRequest []byte
	jsonStringFromRequest, err = definition.ToJson()

	logger.Log.Debug("server received container object", zap.String("definition", string(jsonStringFromRequest)))

	var dr *replicas.Replicas

	obj, err = request.Definition.Apply(format, obj, static.KIND_CONTAINER)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err), err
	} else {
		dr, err = GenerateContainers(container.Shared, user, agent, definition, obj.GetDiff())

		if err != nil {
			return common.Response(http.StatusInternalServerError, "failed to generate replica counts", err), err
		}
	}

	if len(dr.Distributed.Replicas[container.Shared.Manager.Config.KVStore.Node].Remove) > 0 {
		for _, containerObj := range dr.DeleteScoped {
			GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.GetGroup(), containerObj.GetGeneratedName())

			format = f.New("configuration", containerObj.GetGroup(), containerObj.GetGeneratedName(), "")
			obj.Remove(format)

			containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PENDING_DELETE)
			reconcile.Container(container.Shared, container.Shared.Watcher.Find(GroupIdentifier))
		}
	}

	if len(dr.Distributed.Replicas[container.Shared.Manager.Config.KVStore.Node].Create) > 0 {
		for _, containerObj := range dr.CreateScoped {
			GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.GetGroup(), containerObj.GetGeneratedName())
			existingWatcher := container.Shared.Watcher.Find(GroupIdentifier)

			if obj.Exists() {
				if obj.ChangeDetected() {
					container.Shared.Registry.AddOrUpdate(containerObj.GetGroup(), containerObj.GetGeneratedName(), containerObj)

					existingWatcher = reconcile.NewWatcher(containerObj, container.Shared.Manager, user)
					existingWatcher.Logger.Info("container object modified")

					existingWatcher.Container.GetStatus().SetState(status.STATUS_CREATED)
					go reconcile.HandleTickerAndEvents(container.Shared, existingWatcher)

					container.Shared.Watcher.AddOrUpdate(GroupIdentifier, existingWatcher)

					reconcile.Container(container.Shared, existingWatcher)
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

				reconcile.Container(container.Shared, existingWatcher)
			}
		}
	}

	return common.Response(http.StatusOK, "object applied", nil), nil
}
func (container *Container) Compare(user *authentication.User, jsonData []byte) (contracts.Response, error) {
	return common.Response(http.StatusOK, "object in sync", nil), nil
}
func (container *Container) Delete(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequest(static.KIND_CONTAINER)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err), err
	}

	if err = request.Definition.FromJson(jsonData); err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err), err
	}

	definition := request.Definition.Definition.(*v1.ContainerDefinition)

	format := f.New("container", definition.Meta.Group, definition.Meta.Name, "object")
	obj := objects.New(container.Shared.Client.Get(user.Username), user)

	var dr *replicas.Replicas

	existingDefinition, err := request.Definition.Delete(format, obj, static.KIND_CONTAINER)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err), err
	} else {
		dr, err = GetContainers(container.Shared, user, agent, existingDefinition.(*v1.ContainerDefinition))

		if err != nil {
			return common.Response(http.StatusInternalServerError, "failed to generate replica counts", err), err
		}
	}

	if err == nil {
		if len(dr.Distributed.Replicas[container.Shared.Manager.Config.KVStore.Node].Numbers.Existing) > 0 {
			containerObjs := FetchContainersFromRegistry(container.Shared.Registry, dr.Distributed.Replicas[container.Shared.Manager.Config.KVStore.Node].Existing)

			for _, containerObj := range containerObjs {
				go func() {
					GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.GetGroup(), containerObj.GetGeneratedName())
					containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PENDING_DELETE)

					reconcile.Container(container.Shared, container.Shared.Watcher.Find(GroupIdentifier))
				}()
			}

			dr.Distributed.Remove(container.Shared.Client.Clients[user.Username], user)

			return contracts.Response{
				HttpStatus:       200,
				Explanation:      "container is deleted",
				ErrorExplanation: "",
				Error:            false,
				Success:          true,
			}, nil
		} else {
			return common.Response(http.StatusNotFound, "container is not found on the server definition sent", errors.New("container not found")), errors.New("container not found")
		}
	} else {
		return common.Response(http.StatusNotFound, "container is not found on the server definition sent", errors.New("container not found")), errors.New("container not found")

	}
}
func (container *Container) Run(operation string, request contracts.Control) contracts.Response {
	reflected := reflect.TypeOf(container)
	reflectedValue := reflect.ValueOf(container)

	for i := 0; i < reflected.NumMethod(); i++ {
		method := reflected.Method(i)

		if operation == strings.ToLower(method.Name) {
			inputs := []reflect.Value{reflect.ValueOf(request)}
			returnValue := reflectedValue.MethodByName(method.Name).Call(inputs)

			return returnValue[0].Interface().(contracts.Response)
		}
	}

	return common.Response(http.StatusBadRequest, "server doesn't support requested functionality", errors.New("implementation is missing"))
}

func FetchContainersFromRegistry(registry *registry.Registry, containers []replicas.R) []platforms.IContainer {
	var order []platforms.IContainer

	registry.ContainersLock.RLock()
	for _, r := range containers {
		if registry.Containers[r.Group] != nil {
			if registry.Containers[r.Group][r.Name] != nil {
				order = append(order, registry.Containers[r.Group][r.Name])
			}
		}
	}
	registry.ContainersLock.RUnlock()

	return order
}

func GenerateContainers(shared *shared.Shared, user *authentication.User, agent string, containerDefinition *v1.ContainerDefinition, changelog diff.Changelog) (*replicas.Replicas, error) {
	replication := replicas.NewReplica(shared, agent, containerDefinition, changelog)
	err := replication.HandleReplica(user, shared.Manager.Cluster.Cluster.ToString())

	return replication, err
}

func GetContainers(shared *shared.Shared, user *authentication.User, agent string, containerDefinition *v1.ContainerDefinition) (*replicas.Replicas, error) {
	replication := replicas.NewReplica(shared, agent, containerDefinition, nil)
	err := replication.GetReplica(user, shared.Manager.Cluster.Cluster.ToString())

	return replication, err
}

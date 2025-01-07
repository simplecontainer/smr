package container

import (
	"encoding/json"
	"fmt"
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
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
	"github.com/simplecontainer/smr/pkg/raft"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"net/http"
	"reflect"
	"strings"
)

func (container *Container) Start() error {
	container.Started = true

	container.Shared.Watcher = &watcher.ContainerWatcher{
		EventChannel: make(chan raft.KV),
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
	request := NewRequest()

	if err := json.Unmarshal(jsonData, request.Definition); err != nil {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	_, err := request.Definition.Validate()

	if err != nil {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	var format *f.Format
	format = f.New("container", request.Definition.Meta.Group, request.Definition.Meta.Name, "object")

	obj := objects.New(container.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest []byte
	jsonStringFromRequest, err = request.Definition.ToJson()

	logger.Log.Debug("server received container object", zap.String("definition", string(jsonStringFromRequest)))

	var dr *replicas.Replicas

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			dr, err = GenerateContainers(container.Shared, user, agent, request.Definition, obj.Changelog)
			err = obj.Update(format, jsonStringFromRequest)

			if err != nil {
				return contracts.Response{
					HttpStatus:       http.StatusInternalServerError,
					Explanation:      "failed to update object",
					ErrorExplanation: err.Error(),
					Error:            false,
					Success:          true,
				}, err
			}
		} else {
			return contracts.Response{
				HttpStatus:       http.StatusOK,
				Explanation:      "object is same",
				ErrorExplanation: "",
				Error:            false,
				Success:          true,
			}, err
		}
	} else {
		dr, err = GenerateContainers(container.Shared, user, agent, request.Definition, obj.Changelog)

		if err != nil {
			return contracts.Response{
				HttpStatus:       http.StatusInternalServerError,
				Explanation:      "failed to update object",
				ErrorExplanation: err.Error(),
				Error:            false,
				Success:          true,
			}, err
		}

		err = obj.Add(format, jsonStringFromRequest)

		if err != nil {
			return contracts.Response{
				HttpStatus:       http.StatusInternalServerError,
				Explanation:      "failed to add object",
				ErrorExplanation: err.Error(),
				Error:            false,
				Success:          true,
			}, err
		}
	}

	if err == nil {
		fmt.Println(dr.Distributed.Replicas[container.Shared.Manager.Config.KVStore.Node].Create)
		fmt.Println(dr.Distributed.Replicas[container.Shared.Manager.Config.KVStore.Node].Remove)

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

		return contracts.Response{
			HttpStatus:       200,
			Explanation:      "everything went well: good job!",
			ErrorExplanation: "",
			Error:            false,
			Success:          true,
		}, nil
	} else {
		logger.Log.Error(err.Error())

		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "failed to add container",
			ErrorExplanation: err.Error(),
			Error:            false,
			Success:          true,
		}, nil
	}
}
func (container *Container) Compare(user *authentication.User, jsonData []byte) (contracts.Response, error) {
	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "object in sync",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}
func (container *Container) Delete(user *authentication.User, jsonData []byte, agent string) (contracts.Response, error) {
	request := NewRequest()

	if err := json.Unmarshal(jsonData, request.Definition); err != nil {
		return contracts.Response{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	data := make(map[string]interface{})
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		panic(err)
	}

	var format *f.Format
	format = f.New("container", request.Definition.Meta.Group, request.Definition.Meta.Name, "object")

	obj := objects.New(container.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	if obj.Exists() {
		var dr *replicas.Replicas
		dr, err = GetContainers(container.Shared, user, agent, request.Definition)

		if err == nil {
			if len(dr.Distributed.Replicas[container.Shared.Manager.Config.KVStore.Node].Numbers.Existing) > 0 {
				containerObjs := FetchContainersFromRegistry(container.Shared.Registry, dr.Distributed.Replicas[container.Shared.Manager.Config.KVStore.Node].Existing)

				format = f.New("container", request.Definition.Meta.Group, request.Definition.Meta.Name, "")
				obj.Remove(format)

				for _, containerObj := range containerObjs {
					GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.GetGroup(), containerObj.GetGeneratedName())
					containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PENDING_DELETE)
					reconcile.Container(container.Shared, container.Shared.Watcher.Find(GroupIdentifier))
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
				return contracts.Response{
					HttpStatus:       404,
					Explanation:      "",
					ErrorExplanation: "container is not found on the server",
					Error:            true,
					Success:          false,
				}, nil
			}
		} else {
			return contracts.Response{
				HttpStatus:       404,
				Explanation:      "",
				ErrorExplanation: "container is not found on the server",
				Error:            true,
				Success:          false,
			}, nil
		}
	} else {
		return contracts.Response{
			HttpStatus:       404,
			Explanation:      "",
			ErrorExplanation: "container is not found on the server",
			Error:            true,
			Success:          false,
		}, nil
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

	return contracts.Response{
		HttpStatus:       400,
		Explanation:      "server doesn't support requested functionality",
		ErrorExplanation: "implementation is missing",
		Error:            true,
		Success:          false,
		Data:             nil,
	}
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

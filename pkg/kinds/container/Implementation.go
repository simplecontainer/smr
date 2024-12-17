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
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
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
)

func (container *Container) Start() error {
	container.Started = true

	container.Shared.Watcher = &watcher.ContainerWatcher{
		EventChannel: make(chan *types.Events),
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
		panic(err)
	}

	var format *f.Format
	format = f.New("container", request.Definition.Meta.Group, request.Definition.Meta.Name, "object")

	obj := objects.New(container.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = request.Definition.ToJsonString()

	logger.Log.Debug("server received container object", zap.String("definition", jsonStringFromRequest))

	var dr *replicas.Distributed

	if obj.Exists() {
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

		if obj.Diff(jsonStringFromRequest) {
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
		if len(dr.Replicas[container.Shared.Manager.Config.KVStore.Node].Remove) > 0 {
			containerObjs := FetchContainersFromRegistry(container.Shared.Registry, dr.Replicas[container.Shared.Manager.Config.KVStore.Node].Remove)

			for _, containerObj := range containerObjs {
				GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.GetGroup(), containerObj.GetGeneratedName())

				format = f.New("container", containerObj.GetGroup(), containerObj.GetName(), "")
				obj.Remove(format)

				format = f.New("configuration", containerObj.GetGroup(), containerObj.GetGeneratedName(), "")
				obj.Remove(format)

				containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PENDING_DELETE)
				reconcile.Container(container.Shared, container.Shared.Watcher.Find(GroupIdentifier))
			}
		}

		if len(dr.Replicas[container.Shared.Manager.Config.KVStore.Node].Create) > 0 {
			containerObjs := FetchContainersFromRegistry(container.Shared.Registry, dr.Replicas[container.Shared.Manager.Config.KVStore.Node].Create)

			for k, containerObj := range containerObjs {
				// containerObj is fetched from the registry and can be used instead of the object
				// registry is modified by the generateReplicaNamesAndGroups so it is safe to use
				// it instead of the object

				GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.GetGroup(), containerObj.GetGeneratedName())
				containerFromDefinition := container.Shared.Watcher.Find(GroupIdentifier)

				if obj.Exists() {
					if obj.ChangeDetected() || containerFromDefinition == nil {
						if containerFromDefinition == nil {
							containerFromDefinition = reconcile.NewWatcher(containerObjs[k], container.Shared.Manager, user)
							containerFromDefinition.Logger.Info("container object recreated")

							containerFromDefinition.Container.GetStatus().SetState(status.STATUS_RECREATED)
							go reconcile.HandleTickerAndEvents(container.Shared, containerFromDefinition)
						} else {
							containerFromDefinition.Container = containerObjs[k]
							containerFromDefinition.Logger.Info("container object modified")
							containerFromDefinition.Container.GetStatus().SetState(status.STATUS_CREATED)
						}

						container.Shared.Watcher.AddOrUpdate(GroupIdentifier, containerFromDefinition)

						reconcile.Container(container.Shared, containerFromDefinition)
					} else {
						logger.Log.Debug("no change detected in the containers definition")
					}
				} else {
					containerFromDefinition = reconcile.NewWatcher(containerObjs[k], container.Shared.Manager, user)
					containerFromDefinition.Logger.Info("container object created")

					go reconcile.HandleTickerAndEvents(container.Shared, containerFromDefinition)

					containerFromDefinition.Container.GetStatus().SetState(status.STATUS_CREATED)
					container.Shared.Watcher.AddOrUpdate(GroupIdentifier, containerFromDefinition)

					reconcile.Container(container.Shared, containerFromDefinition)
				}
			}
		}
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

	return contracts.Response{
		HttpStatus:       200,
		Explanation:      "everything went well: good job!",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
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
		var dr *replicas.Distributed
		dr, err = GetContainers(container.Shared, user, agent, request.Definition)

		if err == nil {
			if len(dr.Replicas[container.Shared.Manager.Config.KVStore.Node].Numbers.Existing) > 0 {
				containerObjs := FetchContainersFromRegistry(container.Shared.Registry, dr.Replicas[container.Shared.Manager.Config.KVStore.Node].Existing)

				format = f.New("container", request.Definition.Meta.Group, request.Definition.Meta.Name, "")
				obj.Remove(format)

				for _, containerObj := range containerObjs {
					GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.GetGroup(), containerObj.GetGeneratedName())
					containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PENDING_DELETE)
					reconcile.Container(container.Shared, container.Shared.Watcher.Find(GroupIdentifier))
				}

				dr.Remove(container.Shared.Client.Clients[user.Username], user)

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

		if operation == method.Name {
			inputs := []reflect.Value{reflect.ValueOf(request)}
			returnValue := reflectedValue.MethodByName(operation).Call(inputs)

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

// TODO: refactor
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

func GenerateContainers(shared *shared.Shared, user *authentication.User, agent string, containerDefinition *v1.ContainerDefinition, changelog diff.Changelog) (*replicas.Distributed, error) {
	replication := replicas.NewReplica(shared, agent, containerDefinition, changelog)
	return replication.HandleReplica(user, shared.Manager.Cluster.Cluster.ToString())
}

func GetContainers(shared *shared.Shared, user *authentication.User, agent string, containerDefinition *v1.ContainerDefinition) (*replicas.Distributed, error) {
	replication := replicas.NewReplica(shared, agent, containerDefinition, nil)
	return replication.GetReplica(user, shared.Manager.Cluster.Cluster.ToString())
}

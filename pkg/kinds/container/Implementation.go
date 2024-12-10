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
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"net/http"
	"reflect"
)

func (container *Container) Start() error {
	container.Started = true

	container.Shared.Watcher = &watcher.ContainerWatcher{}
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
func (container *Container) Apply(user *authentication.User, jsonData []byte) (contracts.ResponseImplementation, error) {
	request := NewRequest()

	if err := json.Unmarshal(jsonData, request.Definition); err != nil {
		return contracts.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	_, err := request.Definition.Validate()

	if err != nil {
		return contracts.ResponseImplementation{
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

	var dr *replicas.DistributedReplicas

	if obj.Exists() {
		dr, err = generateReplicaNamesAndGroups(container.Shared, user, obj.ChangeDetected(), request.Definition, obj.Changelog)

		if err != nil {
			return contracts.ResponseImplementation{
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
				return contracts.ResponseImplementation{
					HttpStatus:       http.StatusInternalServerError,
					Explanation:      "failed to update object",
					ErrorExplanation: err.Error(),
					Error:            false,
					Success:          true,
				}, err
			}
		}
	} else {
		dr, err = generateReplicaNamesAndGroups(container.Shared, user, obj.ChangeDetected(), request.Definition, obj.Changelog)

		if err != nil {
			return contracts.ResponseImplementation{
				HttpStatus:       http.StatusInternalServerError,
				Explanation:      "failed to update object",
				ErrorExplanation: err.Error(),
				Error:            false,
				Success:          true,
			}, err
		}

		err = obj.Add(format, jsonStringFromRequest)

		if err != nil {
			return contracts.ResponseImplementation{
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

		return contracts.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "failed to add container",
			ErrorExplanation: err.Error(),
			Error:            false,
			Success:          true,
		}, nil
	}

	return contracts.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      "everything went well: good job!",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}
func (container *Container) Compare(user *authentication.User, jsonData []byte) (contracts.ResponseImplementation, error) {
	return contracts.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      "object in sync",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}
func (container *Container) Delete(user *authentication.User, jsonData []byte) (contracts.ResponseImplementation, error) {
	request := NewRequest()

	if err := json.Unmarshal(jsonData, request.Definition); err != nil {
		return contracts.ResponseImplementation{
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
		var containers []replicas.R
		containers, err = GetReplicaNamesAndGroups(container.Shared, user, *request.Definition)

		if err == nil {
			if len(containers) > 0 {
				containerObjs := FetchContainersFromRegistry(container.Shared.Registry, containers)

				format = f.New("container", request.Definition.Meta.Group, request.Definition.Meta.Name, "")
				obj.Remove(format)

				for _, c := range containers {
					format = f.New("configuration", c.Group, c.Name, "")
					obj.Remove(format)
				}

				for _, containerObj := range containerObjs {
					GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.GetGroup(), containerObj.GetGeneratedName())
					containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_PENDING_DELETE)
					reconcile.Container(container.Shared, container.Shared.Watcher.Find(GroupIdentifier))
				}

				return contracts.ResponseImplementation{
					HttpStatus:       200,
					Explanation:      "container is deleted",
					ErrorExplanation: "",
					Error:            false,
					Success:          true,
				}, nil
			} else {
				return contracts.ResponseImplementation{
					HttpStatus:       404,
					Explanation:      "",
					ErrorExplanation: "container is not found on the server",
					Error:            true,
					Success:          false,
				}, nil
			}
		} else {
			return contracts.ResponseImplementation{
				HttpStatus:       404,
				Explanation:      "",
				ErrorExplanation: "container is not found on the server",
				Error:            true,
				Success:          false,
			}, nil
		}
	} else {
		return contracts.ResponseImplementation{
			HttpStatus:       404,
			Explanation:      "",
			ErrorExplanation: "container is not found on the server",
			Error:            true,
			Success:          false,
		}, nil
	}
}
func (container *Container) Run(operation string, args ...interface{}) contracts.ResponseOperator {
	reflected := reflect.TypeOf(container)
	reflectedValue := reflect.ValueOf(container)

	for i := 0; i < reflected.NumMethod(); i++ {
		method := reflected.Method(i)

		if operation == method.Name {
			inputs := make([]reflect.Value, len(args))

			for i, _ := range args {
				inputs[i] = reflect.ValueOf(args[i])
			}

			returnValue := reflectedValue.MethodByName(operation).Call(inputs)

			return returnValue[0].Interface().(contracts.ResponseOperator)
		}
	}

	return contracts.ResponseOperator{
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

	for _, r := range containers {
		if registry.Containers[r.Group] != nil {
			if registry.Containers[r.Group][r.Name] != nil {
				order = append(order, registry.Containers[r.Group][r.Name])
			}
		}
	}

	return order
}

func generateReplicaNamesAndGroups(shared *shared.Shared, user *authentication.User, changed bool, containerDefinition *v1.ContainerDefinition, changelog diff.Changelog) (*replicas.DistributedReplicas, error) {
	_, indexes := shared.Registry.Name(shared.Client, containerDefinition.Meta.Group, containerDefinition.Meta.Name)

	r := replicas.Replicas{
		Group:           containerDefinition.Meta.Group,
		ExistingIndexes: indexes,
		Replicas:        containerDefinition.Spec.Container.Replicas,
		Changed:         changed,
		Spread:          containerDefinition.Spec.Container.Spread,
		NodeID:          shared.Manager.Config.KVStore.Node,
	}

	dr, err := r.HandleReplica(shared, user, shared.Manager.Cluster.Cluster, containerDefinition, changelog)

	return dr, err
}
func GetReplicaNamesAndGroups(shared *shared.Shared, user *authentication.User, containerDefinition v1.ContainerDefinition) ([]replicas.R, error) {
	_, indexes := shared.Registry.Name(shared.Client, containerDefinition.Meta.Group, containerDefinition.Meta.Name)

	r := replicas.Replicas{
		Group:           containerDefinition.Meta.Group,
		ExistingIndexes: indexes,
		Replicas:        containerDefinition.Spec.Container.Replicas,
		NodeID:          shared.Manager.Config.KVStore.Node,
	}

	containers, err := r.GetReplica(shared, user, containerDefinition, shared.Manager.Cluster.Cluster)

	return containers, err
}

func (container *Container) ListSupported(args ...interface{}) contracts.ResponseOperator {
	reflected := reflect.TypeOf(container)

	supportedOperations := map[string]any{}
	supportedOperations["SupportedOperations"] = []string{}

OUTER:
	for i := 0; i < reflected.NumMethod(); i++ {
		method := reflected.Method(i)
		for _, forbiddenOperator := range invalidOperators {
			if forbiddenOperator == method.Name {
				continue OUTER
			}
		}

		supportedOperations["SupportedOperations"] = append(supportedOperations["SupportedOperations"].([]string), method.Name)
	}

	return contracts.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             supportedOperations,
	}
}

func (container *Container) List(request contracts.RequestOperator) contracts.ResponseOperator {
	data := make(map[string]any)

	format := f.New(KIND, "", "", "")

	obj := objects.New(container.Shared.Client.Get(request.User.Username), request.User)
	objs, err := obj.FindMany(format)

	if err != nil {
		return contracts.ResponseOperator{
			HttpStatus:       400,
			Explanation:      "error occured",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	for k, v := range objs {
		data[k] = v.GetDefinition()
	}

	return contracts.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "list of the certkey objects",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             data,
	}
}
func (container *Container) Get(request contracts.RequestOperator) contracts.ResponseOperator {
	if request.Data == nil {
		return contracts.ResponseOperator{
			HttpStatus:       400,
			Explanation:      "send some data",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	format := f.NewFromString(fmt.Sprintf("%s.%s.%s.%s", KIND, request.Data["group"], request.Data["identifier"], "object"))

	obj := objects.New(container.Shared.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)

	if err != nil {
		return contracts.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "container definition is not found on the server",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	definitionObject := obj.GetDefinition()

	var definition = make(map[string]any)
	definition["kind"] = KIND
	definition[KIND] = definitionObject

	return contracts.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "container object is found on the server",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             definition,
	}
}
func (container *Container) View(request contracts.RequestOperator) contracts.ResponseOperator {
	if request.Data == nil {
		return contracts.ResponseOperator{
			HttpStatus:       400,
			Explanation:      "send some data",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	containerObj := container.Shared.Registry.FindLocal(fmt.Sprintf("%s", request.Data["group"]), fmt.Sprintf("%s", request.Data["identifier"]))

	if containerObj == nil {
		return contracts.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "container not found in the registry",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	var definition = make(map[string]any)
	definition[containerObj.GetGeneratedName()] = containerObj

	return contracts.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "container object is found on the server",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             definition,
	}
}
func (container *Container) Restart(request contracts.RequestOperator) contracts.ResponseOperator {
	if request.Data == nil {
		return contracts.ResponseOperator{
			HttpStatus:       400,
			Explanation:      "send some data",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	containerObj := container.Shared.Registry.FindLocal(fmt.Sprintf("%s", request.Data["group"]), fmt.Sprintf("%s", request.Data["identifier"]))

	if containerObj == nil {
		return contracts.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "container not found in the registry",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	containerObj.GetStatus().TransitionState(containerObj.GetGroup(), containerObj.GetGeneratedName(), status.STATUS_CREATED)
	container.Shared.Watcher.Find(containerObj.GetGroupIdentifier()).ContainerQueue <- containerObj

	return contracts.ResponseOperator{
		HttpStatus:       200,
		Explanation:      "container object is restarted",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             nil,
	}
}
func (container *Container) Remove(request contracts.RequestOperator) contracts.ResponseOperator {
	if request.Data == nil {
		return contracts.ResponseOperator{
			HttpStatus:       400,
			Explanation:      "send some data",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	format := f.New("container", request.Data["group"].(string), request.Data["identifier"].(string), "object")

	obj := objects.New(container.Shared.Client.Get(request.User.Username), request.User)
	err := obj.Find(format)
	if err != nil {
		panic(err)
	}

	if !obj.Exists() {
		return contracts.ResponseOperator{
			HttpStatus:       404,
			Explanation:      "object not found on the server",
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
		}
	}

	_, err = container.Delete(request.User, obj.GetDefinitionByte())

	if err != nil {
		return contracts.ResponseOperator{
			HttpStatus:       http.StatusInternalServerError,
			Explanation:      "failed to delete containers",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}

	return contracts.ResponseOperator{
		HttpStatus:       http.StatusOK,
		Explanation:      "action completed successfully",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}
}

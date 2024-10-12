package main

import "C"
import (
	"encoding/json"
	"fmt"
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/implementations/container/container"
	"github.com/simplecontainer/smr/implementations/container/events"
	"github.com/simplecontainer/smr/implementations/container/reconcile"
	"github.com/simplecontainer/smr/implementations/container/registry"
	"github.com/simplecontainer/smr/implementations/container/replicas"
	"github.com/simplecontainer/smr/implementations/container/shared"
	"github.com/simplecontainer/smr/implementations/container/status"
	"github.com/simplecontainer/smr/implementations/container/watcher"
	hubShared "github.com/simplecontainer/smr/implementations/hub/shared"
	"github.com/simplecontainer/smr/pkg/authentication"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/plugins"
	"go.uber.org/zap"
	"net/http"
)

func (implementation *Implementation) Start(mgr *manager.Manager) error {
	implementation.Shared.Manager = mgr
	implementation.Started = true

	container.IsDockerRunning()

	implementation.Shared.Client = mgr.Http
	implementation.Shared.Watcher = &watcher.ContainerWatcher{}
	implementation.Shared.Watcher.Container = make(map[string]*watcher.Container)

	implementation.Shared.Registry = &registry.Registry{
		Containers:     make(map[string]map[string]*container.Container),
		Indexes:        make(map[string][]int),
		BackOffTracker: make(map[string]map[string]int),
	}

	implementation.Shared.DnsCache = mgr.DnsCache

	go events.ListenDockerEvents(implementation.Shared)

	pl := plugins.GetPlugin(implementation.Shared.Manager.Config.OptRoot, "hub.so")
	sharedContainer := pl.GetShared().(*hubShared.Shared)

	go events.ListenEvents(implementation.Shared, sharedContainer.Event)

	return nil
}

func (implementation *Implementation) GetShared() interface{} {
	return implementation.Shared
}

func (implementation *Implementation) Apply(user *authentication.User, jsonData []byte) (httpcontract.ResponseImplementation, error) {
	containerDefinition := &v1.ContainerDefinition{}

	if err := json.Unmarshal(jsonData, &containerDefinition); err != nil {
		return httpcontract.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	_, err := containerDefinition.Validate()

	if err != nil {
		return httpcontract.ResponseImplementation{
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
	format = f.New("container", containerDefinition.Meta.Group, containerDefinition.Meta.Name, "object")

	obj := objects.New(implementation.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = containerDefinition.ToJsonString()

	logger.Log.Debug("server received container object", zap.String("definition", jsonStringFromRequest))

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(format, jsonStringFromRequest)

			if err != nil {
				return httpcontract.ResponseImplementation{
					HttpStatus:       200,
					Explanation:      "failed to update object",
					ErrorExplanation: err.Error(),
					Error:            false,
					Success:          true,
				}, err
			}
		}
	} else {
		err = obj.Add(format, jsonStringFromRequest)

		if err != nil {
			return httpcontract.ResponseImplementation{
				HttpStatus:       http.StatusInternalServerError,
				Explanation:      "failed to add object",
				ErrorExplanation: err.Error(),
				Error:            false,
				Success:          true,
			}, err
		}
	}

	var create map[string][]string
	var remove map[string][]string

	create, remove, err = generateReplicaNamesAndGroups(implementation.Shared, obj.ChangeDetected(), *containerDefinition, obj.Changelog)

	if err == nil {
		if len(remove["groups"]) > 0 {
			containerObjs := FetchContainersFromRegistry(implementation.Shared.Registry, remove["groups"], remove["names"])

			for _, containerObj := range containerObjs {
				GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)

				format = f.New("container", containerObj.Static.Group, containerObj.Static.Name, "")
				obj.Remove(format)

				format = f.New("configuration", containerObj.Static.Group, containerObj.Static.GeneratedName, "")
				obj.Remove(format)

				containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_PENDING_DELETE)
				reconcile.Container(implementation.Shared, implementation.Shared.Watcher.Find(GroupIdentifier))
			}
		}

		if len(create["groups"]) > 0 {
			containerObjs := FetchContainersFromRegistry(implementation.Shared.Registry, create["groups"], create["names"])

			for k, containerObj := range containerObjs {
				// containerObj is fetched from the registry and can be used instead of the object
				// registry is modified by the generateReplicaNamesAndGroups so it is safe to use
				// it instead of the object

				GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)
				containerFromDefinition := implementation.Shared.Watcher.Find(GroupIdentifier)

				if obj.Exists() {
					if obj.ChangeDetected() || containerFromDefinition == nil {
						if containerFromDefinition == nil {
							containerFromDefinition = reconcile.NewWatcher(containerObjs[k], implementation.Shared.Manager, user)
							containerFromDefinition.Logger.Info("container object recreated")

							containerFromDefinition.Container.Status.SetState(status.STATUS_RECREATED)
							go reconcile.HandleTickerAndEvents(implementation.Shared, containerFromDefinition)
						} else {
							containerFromDefinition.Container = containerObjs[k]
							containerFromDefinition.Logger.Info("container object modified")
							containerFromDefinition.Container.Status.SetState(status.STATUS_CREATED)
						}

						implementation.Shared.Watcher.AddOrUpdate(GroupIdentifier, containerFromDefinition)

						reconcile.Container(implementation.Shared, containerFromDefinition)
					} else {
						logger.Log.Debug("no change detected in the containers definition")
					}
				} else {
					containerFromDefinition = reconcile.NewWatcher(containerObjs[k], implementation.Shared.Manager, user)
					containerFromDefinition.Logger.Info("container object created")

					go reconcile.HandleTickerAndEvents(implementation.Shared, containerFromDefinition)

					containerFromDefinition.Container.Status.SetState(status.STATUS_CREATED)
					implementation.Shared.Watcher.AddOrUpdate(GroupIdentifier, containerFromDefinition)

					reconcile.Container(implementation.Shared, containerFromDefinition)
				}
			}
		}
	} else {
		logger.Log.Error(err.Error())

		return httpcontract.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "failed to add container",
			ErrorExplanation: err.Error(),
			Error:            false,
			Success:          true,
		}, nil
	}

	return httpcontract.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      "everything went well: good job!",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}

func (implementation *Implementation) Compare(user *authentication.User, jsonData []byte) (httpcontract.ResponseImplementation, error) {
	return httpcontract.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      "object in sync",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}

func (implementation *Implementation) Delete(user *authentication.User, jsonData []byte) (httpcontract.ResponseImplementation, error) {
	containersDefinition := &v1.ContainerDefinition{}

	if err := json.Unmarshal(jsonData, &containersDefinition); err != nil {
		return httpcontract.ResponseImplementation{
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
	format = f.New("container", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New(implementation.Shared.Client.Get(user.Username), user)
	err = obj.Find(format)

	if obj.Exists() {
		var groups []string
		var names []string

		groups, names, err = GetReplicaNamesAndGroups(implementation.Shared, *containersDefinition)

		if err == nil {
			if len(groups) > 0 {
				containerObjs := FetchContainersFromRegistry(implementation.Shared.Registry, groups, names)

				format = f.New("container", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "")
				obj.Remove(format)

				for k, name := range names {
					format = f.New("configuration", groups[k], name, "")
					obj.Remove(format)
				}

				for _, containerObj := range containerObjs {
					GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)
					containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_PENDING_DELETE)
					reconcile.Container(implementation.Shared, implementation.Shared.Watcher.Find(GroupIdentifier))
				}

				return httpcontract.ResponseImplementation{
					HttpStatus:       200,
					Explanation:      "container is deleted",
					ErrorExplanation: "",
					Error:            false,
					Success:          true,
				}, nil
			} else {
				return httpcontract.ResponseImplementation{
					HttpStatus:       404,
					Explanation:      "",
					ErrorExplanation: "container is not found on the server",
					Error:            true,
					Success:          false,
				}, nil
			}
		} else {
			return httpcontract.ResponseImplementation{
				HttpStatus:       404,
				Explanation:      "",
				ErrorExplanation: "container is not found on the server",
				Error:            true,
				Success:          false,
			}, nil
		}
	} else {
		return httpcontract.ResponseImplementation{
			HttpStatus:       404,
			Explanation:      "",
			ErrorExplanation: "container is not found on the server",
			Error:            true,
			Success:          false,
		}, nil
	}
}

func FetchContainersFromRegistry(registry *registry.Registry, groups []string, names []string) []*container.Container {
	var order []*container.Container

	for i, _ := range names {
		if registry.Containers[groups[i]] != nil {
			if registry.Containers[groups[i]][names[i]] != nil {
				order = append(order, registry.Containers[groups[i]][names[i]])
			}
		}
	}

	return order
}

func generateReplicaNamesAndGroups(shared *shared.Shared, changed bool, containerDefinition v1.ContainerDefinition, changelog diff.Changelog) (map[string][]string, map[string][]string, error) {
	_, index := shared.Registry.Name(containerDefinition.Meta.Group, containerDefinition.Meta.Name, shared.Manager.Config.Environment.PROJECT)

	r := replicas.Replicas{
		Group:          containerDefinition.Meta.Group,
		GeneratedIndex: index - 1,
		Replicas:       containerDefinition.Spec.Container.Replicas,
		Changed:        changed,
	}

	create, remove, err := r.HandleReplica(shared, containerDefinition, changelog)

	return create, remove, err
}

func GetReplicaNamesAndGroups(shared *shared.Shared, containerDefinition v1.ContainerDefinition) ([]string, []string, error) {
	_, index := shared.Registry.Name(containerDefinition.Meta.Group, containerDefinition.Meta.Name, shared.Manager.Config.Environment.PROJECT)

	r := replicas.Replicas{
		Group:          containerDefinition.Meta.Group,
		GeneratedIndex: index - 1,
		Replicas:       containerDefinition.Spec.Container.Replicas,
	}

	groups, names, err := r.GetReplica(shared, containerDefinition)

	return groups, names, err
}

// Exported
var Container Implementation = Implementation{
	Started: false,
	Shared:  &shared.Shared{},
}

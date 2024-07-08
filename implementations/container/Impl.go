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
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/plugins"
	"go.uber.org/zap"
)

func (implementation *Implementation) Start(mgr *manager.Manager) error {
	implementation.Shared.Manager = mgr
	implementation.Started = true

	container.IsDockerRunning()

	client, err := manager.GenerateHttpClient(mgr.Keys)

	if err != nil {
		panic(err)
	}

	implementation.Shared.Client = client
	implementation.Shared.Watcher = &watcher.ContainerWatcher{}
	implementation.Shared.Watcher.Container = make(map[string]*watcher.Container)

	implementation.Shared.Registry = &registry.Registry{
		Containers:     make(map[string]map[string]*container.Container),
		Indexes:        make(map[string][]int),
		BackOffTracker: make(map[string]map[string]int),
	}

	implementation.Shared.DnsCache = mgr.DnsCache

	go events.ListenDockerEvents(implementation.Shared)

	pl := plugins.GetPlugin(implementation.Shared.Manager.Config.Root, "hub.so")
	sharedContainer := pl.GetShared().(*hubShared.Shared)

	go events.ListenEvents(implementation.Shared, sharedContainer.Event)

	return nil
}

func (implementation *Implementation) GetShared() interface{} {
	return implementation.Shared
}

func (implementation *Implementation) Apply(jsonData []byte) (httpcontract.ResponseImplementation, error) {
	containerDefinition := &v1.Container{}

	if err := json.Unmarshal(jsonData, &containerDefinition); err != nil {
		return httpcontract.ResponseImplementation{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	}

	valid, err := containerDefinition.Validate()

	if !valid {
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

	var format objects.FormatStructure
	format = objects.Format("container", containerDefinition.Meta.Group, containerDefinition.Meta.Name, "object")

	obj := objects.New()
	err = obj.Find(implementation.Shared.Client, format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = containerDefinition.ToJsonString()

	logger.Log.Debug("server received container object", zap.String("definition", jsonStringFromRequest))

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(implementation.Shared.Client, format, jsonStringFromRequest)
		}
	} else {
		err = obj.Add(implementation.Shared.Client, format, jsonStringFromRequest)
	}

	groups, names, err := generateReplicaNamesAndGroups(implementation.Shared, obj.ChangeDetected(), *containerDefinition, obj.Changelog)

	if err == nil {
		if len(groups) > 0 {
			containerObjs := FetchContainersFromRegistry(implementation.Shared.Registry, groups, names)

			for k, containerObj := range containerObjs {
				if obj.ChangeDetected() || !obj.Exists() {
					containerFromDefinition := reconcile.NewWatcher(containerObjs[k], implementation.Shared.Manager)
					GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)

					containerFromDefinition.Logger.Info("new container object created",
						zap.String("group", containerFromDefinition.Container.Static.Definition.Meta.Group),
						zap.String("identifier", containerFromDefinition.Container.Static.Definition.Meta.Name),
					)

					ContainerTracker := implementation.Shared.Watcher.Find(GroupIdentifier)

					containerFromDefinition.Container.Status.SetState(status.STATUS_CREATED)
					implementation.Shared.Watcher.AddOrUpdate(GroupIdentifier, containerFromDefinition)

					if ContainerTracker == nil {
						go reconcile.HandleTickerAndEvents(implementation.Shared, containerFromDefinition)
					} else {
						if ContainerTracker.Tracking == false {
							ContainerTracker.Tracking = true
							go reconcile.HandleTickerAndEvents(implementation.Shared, containerFromDefinition)
						}
					}

					reconcile.ReconcileContainer(implementation.Shared, containerFromDefinition)
				} else {
					logger.Log.Debug("no change detected in the containers definition")
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

func (implementation *Implementation) Compare(jsonData []byte) (httpcontract.ResponseImplementation, error) {
	return httpcontract.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      "object in sync",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
	}, nil
}

func (implementation *Implementation) Delete(jsonData []byte) (httpcontract.ResponseImplementation, error) {
	containersDefinition := &v1.Container{}

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

	var format objects.FormatStructure
	format = objects.Format("container", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New()
	err = obj.Find(implementation.Shared.Client, format)

	if obj.Exists() {
		groups, names, err := GetReplicaNamesAndGroups(implementation.Shared, *containersDefinition)

		if err == nil {
			if len(groups) > 0 {
				containerObjs := FetchContainersFromRegistry(implementation.Shared.Registry, groups, names)

				for _, containerObj := range containerObjs {
					GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)

					format = objects.Format("container", containerObj.Static.Group, containerObj.Static.Name, "")
					obj.Remove(implementation.Shared.Client, format)

					format = objects.Format("configuration", containerObj.Static.Group, containerObj.Static.GeneratedName, "")
					obj.Remove(implementation.Shared.Client, format)

					containerObj.Status.TransitionState(containerObj.Static.GeneratedName, status.STATUS_PENDING_DELETE)
					reconcile.ReconcileContainer(implementation.Shared, implementation.Shared.Watcher.Find(GroupIdentifier))
				}
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
				Explanation:      "container is not found on the server",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
			}, nil
		}
	}

	return httpcontract.ResponseImplementation{
		HttpStatus:       200,
		Explanation:      "container is not found on the server",
		ErrorExplanation: err.Error(),
		Error:            false,
		Success:          true,
	}, nil
}

func FetchContainersFromRegistry(registry *registry.Registry, groups []string, names []string) []*container.Container {
	var order []*container.Container

	for i, _ := range names {
		order = append(order, registry.Containers[groups[i]][names[i]])
	}

	return order
}

func generateReplicaNamesAndGroups(shared *shared.Shared, changed bool, containerDefinition v1.Container, changelog diff.Changelog) ([]string, []string, error) {
	_, index := shared.Registry.Name(containerDefinition.Meta.Group, containerDefinition.Meta.Name, shared.Manager.Config.Environment.PROJECT)

	r := replicas.Replicas{
		Group:          containerDefinition.Meta.Group,
		GeneratedIndex: index - 1,
		Replicas:       containerDefinition.Spec.Container.Replicas,
		Changed:        changed,
	}

	groups, names, err := r.HandleReplica(shared, containerDefinition, changelog)

	return groups, names, err
}

func GetReplicaNamesAndGroups(shared *shared.Shared, containerDefinition v1.Container) ([]string, []string, error) {
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

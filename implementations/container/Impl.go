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
	"github.com/simplecontainer/smr/pkg/database"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/objects"
	"strconv"
)

func (implementation *Implementation) Start(mgr *manager.Manager) error {
	implementation.Shared.Manager = mgr
	implementation.Started = true

	implementation.Shared.Watcher = &watcher.ContainerWatcher{}
	implementation.Shared.Watcher.Container = make(map[string]*watcher.Container)

	implementation.Shared.Registry = &registry.Registry{
		Containers:     make(map[string]map[string]*container.Container),
		Indexes:        make(map[string][]int),
		BackOffTracker: make(map[string]map[string]int),
	}

	implementation.Shared.DnsCache = &dns.Records{}

	go events.ListenDockerEvents(implementation.Shared)

	return nil
}

func (implementation *Implementation) GetShared() interface{} {
	return implementation.Shared
}

func (implementation *Implementation) Apply(jsonData []byte) (httpcontract.ResponseImplementation, error) {
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

	var format database.FormatStructure
	format = database.Format("container", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New()
	err = obj.Find(implementation.Shared.Manager.Badger, format)

	var jsonStringFromRequest string
	jsonStringFromRequest, err = containersDefinition.ToJsonString()

	if obj.Exists() {
		if obj.Diff(jsonStringFromRequest) {
			err = obj.Update(implementation.Shared.Manager.Badger, format, jsonStringFromRequest)
		}
	} else {
		err = obj.Add(implementation.Shared.Manager.Badger, format, jsonStringFromRequest)
	}

	if obj.ChangeDetected() || !obj.Exists() {
		err = obj.Update(implementation.Shared.Manager.Badger, format, jsonStringFromRequest)
	}

	groups, names, err := generateReplicaNamesAndGroups(implementation.Shared, obj.ChangeDetected(), *containersDefinition, obj.Changelog)

	if err == nil {
		if len(groups) > 0 {
			containerObjs := FetchContainersFromRegistry(implementation.Shared.Registry, groups, names)

			for k, containerObj := range containerObjs {
				if obj.ChangeDetected() || !obj.Exists() {
					containerFromDefinition := reconcile.NewWatcher(containerObjs[k])
					GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)

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
					logger.Log.Info("no change detected in the containers definition")
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
		Explanation:      strconv.Itoa(implementation.State),
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

	var format database.FormatStructure
	format = database.Format("container", containersDefinition.Meta.Group, containersDefinition.Meta.Name, "object")

	obj := objects.New()
	err = obj.Find(implementation.Shared.Manager.Badger, format)

	if obj.Exists() {
		groups, names, err := GetReplicaNamesAndGroups(implementation.Shared, *containersDefinition)

		if err == nil {
			if len(groups) > 0 {
				containerObjs := FetchContainersFromRegistry(implementation.Shared.Registry, groups, names)

				for _, containerObj := range containerObjs {
					GroupIdentifier := fmt.Sprintf("%s.%s", containerObj.Static.Group, containerObj.Static.GeneratedName)

					containerObj.Status.TransitionState(status.STATUS_PENDING_DELETE)

					format = database.Format("runtime", containerObj.Static.Group, containerObj.Static.GeneratedName, "")
					obj.Remove(implementation.Shared.Manager.Badger, format)

					format = database.Format("configuration", containerObj.Static.Group, containerObj.Static.GeneratedName, "")
					obj.Remove(implementation.Shared.Manager.Badger, format)

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
	_, index := shared.Registry.Name(containerDefinition.Meta.Group, containerDefinition.Meta.Name, shared.Manager.Runtime.PROJECT)

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
	_, index := shared.Registry.Name(containerDefinition.Meta.Group, containerDefinition.Meta.Name, shared.Manager.Runtime.PROJECT)

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
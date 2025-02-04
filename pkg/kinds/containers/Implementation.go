package containers

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/containers/reconcile"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
)

func (containers *Containers) Start() error {
	containers.Started = true

	containers.Shared.Watcher = &watcher.ContainersWatcher{}
	containers.Shared.Watcher.Containers = make(map[string]*watcher.Containers)

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

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", request.Definition.GetMeta().Group, request.Definition.GetMeta().Name)
	containersFromDefinition := containers.Shared.Watcher.Find(GroupIdentifier)

	if obj.Exists() {
		if obj.ChangeDetected() || containersFromDefinition == nil {
			if containersFromDefinition == nil {
				containersFromDefinition = reconcile.NewWatcher(*(request.Definition.Definition.(*v1.ContainersDefinition)), containers.Shared.Manager)
				containersFromDefinition.Logger.Info("containers object created")

				go reconcile.HandleTickerAndEvents(containers.Shared, user, containersFromDefinition, agent)
			} else {
				containersFromDefinition.Definition = *(request.Definition.Definition.(*v1.ContainersDefinition))
				containersFromDefinition.Logger.Info("containers object modified")
			}

			containers.Shared.Watcher.AddOrUpdate(GroupIdentifier, containersFromDefinition)
			reconcile.Container(containers.Shared, user, containersFromDefinition, agent)
		} else {
			return common.Response(http.StatusOK, "object applied", nil, nil), nil
		}
	} else {
		containersFromDefinition = reconcile.NewWatcher(*(request.Definition.Definition.(*v1.ContainersDefinition)), containers.Shared.Manager)
		containersFromDefinition.Logger.Info("containers object created")

		go reconcile.HandleTickerAndEvents(containers.Shared, user, containersFromDefinition, agent)

		containers.Shared.Watcher.AddOrUpdate(GroupIdentifier, containersFromDefinition)
		reconcile.Container(containers.Shared, user, containersFromDefinition, agent)
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
		return common.Response(http.StatusInternalServerError, "", err, nil), err
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", request.Definition.GetMeta().Group, request.Definition.GetMeta().Name)

	if containers.Shared.Watcher.Find(GroupIdentifier) != nil {
		containers.Shared.Watcher.Find(GroupIdentifier).Syncing = true
		containers.Shared.Watcher.Find(GroupIdentifier).Cancel()
	}

	spec := (*(request.Definition.Definition.(*v1.ContainersDefinition))).Spec

	for _, container := range spec {
		def, _ := container.ToJson()
		go func() {
			_, err = containers.Shared.Manager.KindsRegistry["container"].Delete(user, def, agent)
			if err != nil {
				logger.Log.Error(err.Error())
			}
		}()
	}

	return common.Response(http.StatusOK, "object deleted", nil, nil), nil
}

func (containers *Containers) Event(event contracts.Event) error {
	return nil
}

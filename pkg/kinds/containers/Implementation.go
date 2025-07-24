package containers

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"github.com/simplecontainer/smr/pkg/contracts/ishared"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/events/platform/listener"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/replicas"
	"github.com/simplecontainer/smr/pkg/kinds/containers/registry"
	"github.com/simplecontainer/smr/pkg/kinds/containers/shared"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/kinds/containers/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/metrics"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/wI2L/jsondiff"
	"net/http"
	"os"
)

func (containers *Containers) Start() error {
	containers.Started = true

	containers.Shared.Watchers = watcher.NewWatchers()
	containers.Shared.Registry = registry.New(containers.Shared.Client, containers.Shared.User)

	logger.Log.Info(fmt.Sprintf("platform for running containers is %s", containers.Shared.Manager.Config.Platform))

	// Check if everything alright with the daemon
	switch containers.Shared.Manager.Config.Platform {
	case static.PLATFORM_DOCKER:
		version, err := docker.IsDaemonRunning()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		metrics.DockerVersion.Increment(version)
		break
	}

	// Start listening events based on the platform and for internal events
	go listener.Listen(containers.Shared, containers.Shared.Manager.Config.Platform)

	logger.Log.Info(fmt.Sprintf("started listening events for simplecontainer and platform: %s", containers.Shared.Manager.Config.Platform))

	return nil
}
func (containers *Containers) GetShared() ishared.Shared {
	return containers.Shared
}
func (containers *Containers) Apply(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CONTAINERS, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	obj, err := request.Apply(containers.Shared.Client, user)

	if request.Definition.GetState() != nil && !request.Definition.GetState().GetOpt("replay").IsEmpty() {
		request.Definition.GetState().ClearOpt("replay")
	} else {
		if !obj.ChangeDetected() {
			return common.Response(http.StatusOK, static.RESPONSE_APPLIED, nil, nil), nil
		}
	}

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	var create []platforms.IContainer
	var update []platforms.IContainer
	var destroy []platforms.IContainer

	create, update, destroy, err = GenerateContainers(containers.Shared, request.Definition.Definition.(*v1.ContainersDefinition), obj.GetDiff())

	if err != nil {
		return common.Response(http.StatusInternalServerError, "failed to generate replica counts", err, nil), err
	}

	if len(destroy) > 0 {
		containers.Destroy(destroy, obj.Exists())
	}

	if len(update) > 0 {
		if len(obj.GetDiff()) == 1 {
			if obj.GetDiff()[0].Path == "/spec/replicas" {
				update = nil
			}
		}

		containers.Update(update, obj.Exists())
	}

	if len(create) > 0 {
		containers.Create(create, obj.Exists(), user)
	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}
func (containers *Containers) State(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CONTAINERS, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.State(containers.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	} else {
		return common.Response(http.StatusOK, "", err, nil), err
	}
}
func (containers *Containers) Delete(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_CONTAINERS, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Remove(containers.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusTeapot, "", err, nil), err
	}

	var destroy []platforms.IContainer
	destroy, err = GetContainers(containers.Shared, request.Definition.Definition.(*v1.ContainersDefinition))

	if err != nil {
		return common.Response(http.StatusInternalServerError, "failed to generate replica counts", err, nil), err
	}

	if len(destroy) > 0 {
		for _, containerObj := range destroy {
			go func() {
				groupIdentifier := containerObj.GetGroupIdentifier()
				containerW := containers.Shared.Watchers.Find(groupIdentifier)

				if containerW != nil && !containerW.Done {
					containers.Shared.Watchers.Find(groupIdentifier).DeleteC <- containerObj
				}
			}()
		}

		return common.Response(http.StatusOK, "object is deleted", nil, nil), nil
	} else {
		return common.Response(http.StatusNotFound, "object not found", errors.New("object not found"), nil), errors.New("object not found")
	}
}
func (containers *Containers) Event(event ievents.Event) error {
	switch event.GetType() {
	case events.EVENT_CHANGE:
		for _, containerWatcher := range containers.Shared.Watchers.Watchers {
			if containerWatcher.Container.HasDependencyOn(event.GetKind(), event.GetGroup(), event.GetName()) {
				err := containerWatcher.Container.GetStatus().QueueState(status.CHANGE)
				if err != nil {
					containerWatcher.Logger.Error(err.Error())
					return err
				}

				containerWatcher.Logger.Info("responding to change")
				containers.Shared.Watchers.Find(containerWatcher.Container.GetGroupIdentifier()).ContainerQueue <- containerWatcher.Container
			}
		}

		return nil
	case events.EVENT_RESTART:
		containerObj := containers.Shared.Registry.FindLocal(event.GetGroup(), event.GetName())

		if containerObj == nil {
			return errors.New("container not found event is ignored")
		}

		containerW := containers.Shared.Watchers.Find(containerObj.GetGroupIdentifier())

		if !containerW.Done {
			containerObj.GetStatus().QueueState(status.RESTART)
			containerW.ContainerQueue <- containerObj
		}

		break
	}

	return nil
}

func GenerateContainers(shared *shared.Shared, definition *v1.ContainersDefinition, changelog jsondiff.Patch) ([]platforms.IContainer, []platforms.IContainer, []platforms.IContainer, error) {
	r := replicas.New(shared.Manager.Cluster.Node.NodeID, shared.Manager.Cluster.Cluster.Nodes)
	return r.GenerateContainers(shared.Registry, definition, shared.Manager.Config)
}
func GetContainers(shared *shared.Shared, definition *v1.ContainersDefinition) ([]platforms.IContainer, error) {
	r := replicas.New(shared.Manager.Cluster.Node.NodeID, shared.Manager.Cluster.Cluster.Nodes)
	return r.RemoveContainers(shared.Registry, definition)
}

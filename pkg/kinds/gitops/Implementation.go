package gitops

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/events/events"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/reconcile"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/registry"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/static"
	"net/http"
)

func (gitops *Gitops) Start() error {
	gitops.Started = true

	gitops.Shared.Watcher = &watcher.RepositoryWatcher{
		Repositories: make(map[string]*watcher.Gitops),
	}

	gitops.Shared.Registry = &registry.Registry{
		Gitopses: make(map[string]map[string]*implementation.Gitops),
		Client:   gitops.Shared.Client,
		User:     gitops.Shared.Manager.User,
	}

	return nil
}
func (gitops *Gitops) GetShared() interface{} {
	return gitops.Shared
}
func (gitops *Gitops) Apply(user *authentication.User, definition []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_GITOPS, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	obj, err := request.Apply(gitops.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", request.Definition.GetMeta().Group, request.Definition.GetMeta().Name)
	gitopsWatcherFromRegistry := gitops.Shared.Watcher.Find(GroupIdentifier)

	if request.Definition.GetRuntime().GetNode() != gitops.Shared.Manager.Cluster.Node.NodeID {
		return common.Response(http.StatusOK, "object applied", nil, nil), nil
	}

	if obj.Exists() {
		if obj.ChangeDetected() || gitopsWatcherFromRegistry == nil {
			if gitopsWatcherFromRegistry == nil {
				gitopsWatcherFromRegistry = reconcile.NewWatcher(implementation.New(request.Definition.Definition.(*v1.GitopsDefinition)), gitops.Shared.Manager, user)
				go reconcile.HandleTickerAndEvents(gitops.Shared, gitopsWatcherFromRegistry)

				gitopsWatcherFromRegistry.Logger.Info("new gitops object created")
				gitopsWatcherFromRegistry.Gitops.Status.SetState(status.STATUS_CREATED)
			} else {
				gitops.Shared.Watcher.Find(GroupIdentifier).Gitops = implementation.New(request.Definition.Definition.(*v1.GitopsDefinition))

				gitopsWatcherFromRegistry.Logger.Info("gitops object modified")
				gitopsWatcherFromRegistry.Gitops.Status.SetState(status.STATUS_CREATED)
			}

			gitops.Shared.Watcher.AddOrUpdate(GroupIdentifier, gitopsWatcherFromRegistry)
			gitops.Shared.Registry.AddOrUpdate(gitopsWatcherFromRegistry.Gitops.GetGroup(), gitopsWatcherFromRegistry.Gitops.GetName(), gitopsWatcherFromRegistry.Gitops)

			reconcile.Gitops(gitops.Shared, gitopsWatcherFromRegistry)
		}
	} else {
		gitopsWatcherFromRegistry = reconcile.NewWatcher(implementation.New(request.Definition.Definition.(*v1.GitopsDefinition)), gitops.Shared.Manager, user)
		go reconcile.HandleTickerAndEvents(gitops.Shared, gitopsWatcherFromRegistry)

		gitopsWatcherFromRegistry.Logger.Info("new gitops object created")
		gitopsWatcherFromRegistry.Gitops.Status.SetState(status.STATUS_CREATED)

		gitops.Shared.Registry.AddOrUpdate(gitopsWatcherFromRegistry.Gitops.GetGroup(), gitopsWatcherFromRegistry.Gitops.GetName(), gitopsWatcherFromRegistry.Gitops)
		gitops.Shared.Watcher.AddOrUpdate(GroupIdentifier, gitopsWatcherFromRegistry)

		reconcile.Gitops(gitops.Shared, gitopsWatcherFromRegistry)
	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}
func (gitops *Gitops) Compare(user *authentication.User, definition []byte) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_GITOPS, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Apply(gitops.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusTeapot, "object drifted", nil, nil), nil
	} else {
		return common.Response(http.StatusOK, "object in sync", nil, nil), nil
	}
}

func (gitops *Gitops) Delete(user *authentication.User, definition []byte, agent string) (contracts.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_GITOPS, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Remove(gitops.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusInternalServerError, "", err, nil), err
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", request.Definition.GetMeta().Group, request.Definition.GetMeta().Name)

	gitopsObj := gitops.Shared.Watcher.Find(GroupIdentifier).Gitops

	gitopsObj.Status.TransitionState(gitopsObj.Definition.Meta.Name, status.STATUS_PENDING_DELETE)
	reconcile.Gitops(gitops.Shared, gitops.Shared.Watcher.Find(GroupIdentifier))

	return common.Response(http.StatusOK, "object in deleted", nil, nil), nil

}

func (gitops *Gitops) Event(event contracts.Event) error {
	switch event.GetType() {
	case events.EVENT_REFRESH:
		gitopsObj := gitops.Shared.Registry.FindLocal(event.GetGroup(), event.GetName())

		if gitopsObj == nil {
			return errors.New("gitops is not controlled on this instance")
		}

		gitopsWatcher := gitops.Shared.Watcher.Find(gitopsObj.GetGroupIdentifier())

		if gitopsWatcher != nil {
			gitopsWatcher.Gitops.ForcePoll = true
			gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONING_GIT)
			gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
		}
		break
	case events.EVENT_SYNC:
		gitopsObj := gitops.Shared.Registry.FindLocal(event.GetGroup(), event.GetName())

		if gitopsObj == nil {
			return errors.New("gitops is not controlled on this instance")
		}

		gitopsWatcher := gitops.Shared.Watcher.Find(gitopsObj.GetGroupIdentifier())

		if gitopsWatcher != nil {
			gitopsWatcher.Gitops.ManualSync = true
			gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONING_GIT)
			gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
		}
		break
	}

	return nil
}

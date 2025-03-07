package gitops

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
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
	"sync"
)

func (gitops *Gitops) Start() error {
	gitops.Started = true

	gitops.Shared.Watcher = &watcher.RepositoryWatcher{
		Repositories: make(map[string]*watcher.Gitops),
		Lock:         &sync.RWMutex{},
	}

	gitops.Shared.Registry = &registry.Registry{
		Gitops: make(map[string]*implementation.Gitops),
		Client: gitops.Shared.Client,
		User:   gitops.Shared.Manager.User,
	}

	return nil
}
func (gitops *Gitops) GetShared() interface{} {
	return gitops.Shared
}
func (gitops *Gitops) Apply(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_GITOPS, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	obj, err := request.Apply(gitops.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	}

	GroupIdentifier := fmt.Sprintf("%s.%s", request.Definition.GetMeta().Group, request.Definition.GetMeta().Name)
	existingWatcher := gitops.Shared.Watcher.Find(GroupIdentifier)

	if request.Definition.GetRuntime().GetNode() != gitops.Shared.Manager.Cluster.Node.NodeID {
		// Only one node can control gitops other will just have copy of the object and state
		return common.Response(http.StatusOK, "object applied", nil, nil), nil
	}

	if obj.Exists() {
		if obj.ChangeDetected() {
			gitopsObj := implementation.New(request.Definition.Definition.(*v1.GitopsDefinition), gitops.Shared.Manager.Config)

			if existingWatcher == nil {
				w := watcher.New(gitopsObj, gitops.Shared.Manager, user)
				go reconcile.HandleTickerAndEvents(gitops.Shared, w, func(w *watcher.Gitops) error {
					return nil
				})

				gitops.Shared.Watcher.AddOrUpdate(GroupIdentifier, w)
				gitops.Shared.Registry.AddOrUpdate(gitopsObj.GetGroup(), gitopsObj.GetName(), gitopsObj)

				w.Gitops.Status.SetState(status.CREATED)
				w.GitopsQueue <- gitopsObj
			} else {
				existingWatcher.Gitops = gitopsObj
				gitops.Shared.Registry.AddOrUpdate(gitopsObj.GetGroup(), gitopsObj.GetName(), gitopsObj)

				existingWatcher.Gitops.Status.SetState(status.CREATED)
				existingWatcher.GitopsQueue <- gitopsObj
			}
		}
	} else {
		gitopsObj := implementation.New(request.Definition.Definition.(*v1.GitopsDefinition), gitops.Shared.Manager.Config)

		w := watcher.New(gitopsObj, gitops.Shared.Manager, user)

		w.Logger.Info("new gitops object created")
		w.Gitops.Status.SetState(status.CREATED)

		gitops.Shared.Registry.AddOrUpdate(w.Gitops.GetGroup(), w.Gitops.GetName(), w.Gitops)
		gitops.Shared.Watcher.AddOrUpdate(GroupIdentifier, w)

		go reconcile.HandleTickerAndEvents(gitops.Shared, w, func(w *watcher.Gitops) error {
			return nil
		})

		w.GitopsQueue <- gitopsObj
	}

	return common.Response(http.StatusOK, "object applied", nil, nil), nil
}
func (gitops *Gitops) State(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_GITOPS, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.State(gitops.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusBadRequest, "", err, nil), err
	} else {
		return common.Response(http.StatusOK, "", err, nil), err
	}
}
func (gitops *Gitops) Delete(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_GITOPS, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	_, err = request.Remove(gitops.Shared.Client, user)

	if err != nil {
		return common.Response(http.StatusInternalServerError, "", err, nil), err
	}

	gitopsObj := gitops.Shared.Registry.FindLocal(request.Definition.GetMeta().Group, request.Definition.GetMeta().Name)

	if gitopsObj == nil {
		return common.Response(http.StatusNotFound, static.RESPONSE_NOT_FOUND, nil, nil), nil
	} else {
		gitopsWatcher := gitops.Shared.Watcher.Find(gitopsObj.GetGroupIdentifier())

		if gitopsWatcher != nil {
			gitopsObj.GetStatus().TransitionState(gitopsWatcher.Gitops.GetGroup(), gitopsWatcher.Gitops.GetName(), status.PENDING_DELETE)
			gitopsWatcher.GitopsQueue <- gitopsObj

			return common.Response(http.StatusOK, static.RESPONSE_DELETED, nil, nil), nil
		} else {
			return common.Response(http.StatusNotFound, static.RESPONSE_NOT_FOUND, nil, nil), nil
		}
	}
}
func (gitops *Gitops) Event(event ievents.Event) error {
	switch event.GetType() {
	case events.EVENT_REFRESH:
		gitopsObj := gitops.Shared.Registry.FindLocal(event.GetGroup(), event.GetName())

		if gitopsObj == nil {
			return errors.New("gitops is not controlled on this instance")
		}

		gitopsWatcher := gitops.Shared.Watcher.Find(gitopsObj.GetGroupIdentifier())

		if gitopsWatcher != nil && !gitopsWatcher.Gitops.Status.PendingDelete {
			gitopsObj.ForcePoll = true
			gitopsObj.GetStatus().SetState(status.CLONING_GIT)
			gitopsWatcher.GitopsQueue <- gitopsObj
		}
		break
	case events.EVENT_SYNC:
		gitopsObj := gitops.Shared.Registry.FindLocal(event.GetGroup(), event.GetName())

		if gitopsObj == nil {
			return errors.New("gitops is not controlled on this instance")
		}

		gitopsWatcher := gitops.Shared.Watcher.Find(gitopsObj.GetGroupIdentifier())

		if gitopsWatcher != nil && !gitopsWatcher.Gitops.Status.PendingDelete {
			gitopsObj.DoSync = true
			gitopsObj.GetStatus().SetState(status.CLONING_GIT)
			gitopsWatcher.GitopsQueue <- gitopsObj
		}
		break
	case events.EVENT_INSPECT:
		gitopsObj := gitops.Shared.Registry.FindLocal(event.GetGroup(), event.GetName())

		if gitopsObj == nil {
			return errors.New("gitops is not controlled on this instance")
		}

		gitopsWatcher := gitops.Shared.Watcher.Find(gitopsObj.GetGroupIdentifier())

		if gitopsWatcher != nil && !gitopsWatcher.Gitops.Status.PendingDelete {
			gitopsObj.GetStatus().SetState(status.INSPECTING)
			gitopsWatcher.GitopsQueue <- gitopsObj
		}
		break
	}

	return nil

}

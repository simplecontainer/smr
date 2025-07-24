package gitops

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"github.com/simplecontainer/smr/pkg/contracts/ishared"
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

	gitops.Shared.Watchers = watcher.NewWatchers()
	gitops.Shared.Registry = registry.New(gitops.Shared.Client, gitops.Shared.Manager.User)

	return nil
}
func (gitops *Gitops) GetShared() ishared.Shared {
	return gitops.Shared
}
func (gitops *Gitops) Apply(user *authentication.User, definition []byte, agent string) (iresponse.Response, error) {
	request, err := common.NewRequestFromJson(static.KIND_GITOPS, definition)

	if err != nil {
		return common.Response(http.StatusBadRequest, "invalid definition sent", err, nil), err
	}

	obj, err := request.Apply(gitops.Shared.Client, user)

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

	GroupIdentifier := fmt.Sprintf("%s/%s", request.Definition.GetMeta().Group, request.Definition.GetMeta().Name)
	existingWatcher := gitops.Shared.Watchers.Find(GroupIdentifier)

	if request.Definition.GetRuntime().GetNode() != gitops.Shared.Manager.Cluster.Node.NodeID {
		// Only one node can control gitops other will just have copy of the object and state
		return common.Response(http.StatusOK, "object applied", nil, nil), nil
	}

	if obj.Exists() {
		gitopsObj, err := implementation.New(request.Definition.Definition.(*v1.GitopsDefinition), gitops.Shared.Manager.Config)

		if err != nil {
			return common.Response(http.StatusInternalServerError, "", err, nil), err
		}

		if existingWatcher == nil {
			w := watcher.New(gitopsObj, gitops.Shared.Manager, user)
			go reconcile.HandleTickerAndEvents(gitops.Shared, w, func(w *watcher.Gitops) error {
				return nil
			})

			gitops.Shared.Watchers.AddOrUpdate(GroupIdentifier, w)
			gitops.Shared.Registry.AddOrUpdate(gitopsObj.GetGroup(), gitopsObj.GetName(), gitopsObj)

			w.Gitops.GetStatus().SetState(status.CREATED)
			w.GitopsQueue <- gitopsObj
		} else {
			existingWatcher.Gitops = gitopsObj
			gitops.Shared.Registry.AddOrUpdate(gitopsObj.GetGroup(), gitopsObj.GetName(), gitopsObj)

			existingWatcher.Gitops.GetStatus().QueueState(status.CREATED)
			existingWatcher.GitopsQueue <- gitopsObj
		}
	} else {
		gitopsObj, err := implementation.New(request.Definition.Definition.(*v1.GitopsDefinition), gitops.Shared.Manager.Config)

		if err != nil {
			return common.Response(http.StatusInternalServerError, "", err, nil), err
		}

		w := watcher.New(gitopsObj, gitops.Shared.Manager, user)

		w.Logger.Info("new gitops object created")
		w.Gitops.GetStatus().QueueState(status.CREATED)

		gitops.Shared.Registry.AddOrUpdate(w.Gitops.GetGroup(), w.Gitops.GetName(), w.Gitops)
		gitops.Shared.Watchers.AddOrUpdate(GroupIdentifier, w)

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
		gitopsWatcher := gitops.Shared.Watchers.Find(gitopsObj.GetGroupIdentifier())

		if gitopsWatcher != nil {
			gitopsObj.GetStatus().QueueState(status.DELETE)
			gitopsWatcher.GitopsQueue <- gitopsObj

			return common.Response(http.StatusOK, static.RESPONSE_DELETED, nil, nil), nil
		} else {
			return common.Response(http.StatusNotFound, static.RESPONSE_NOT_FOUND, nil, nil), nil
		}
	}
}
func (gitops *Gitops) Event(event ievents.Event) error {
	switch event.GetType() {
	case events.EVENT_COMMIT:
		gitopsObj := gitops.Shared.Registry.FindLocal(event.GetGroup(), event.GetName())

		if gitopsObj == nil {
			return nil
		}

		gitopsWatcher := gitops.Shared.Watchers.Find(gitopsObj.GetGroupIdentifier())

		if gitopsWatcher != nil {
			if gitopsWatcher.Gitops.GetStatus().GetPending().Is(status.PENDING_SYNC, status.PENDING_DELETE) {
				return nil
			}

			commit := implementation.NewCommit()

			err := commit.FromJson(event.GetData())

			if err != nil {
				return err
			}

			gitopsObj.GetQueue().Insert(commit)
			gitopsObj.GetStatus().QueueState(status.COMMIT_GIT)
			gitopsWatcher.GitopsQueue <- gitopsObj
		}
		break
	case events.EVENT_REFRESH:
		gitopsObj := gitops.Shared.Registry.FindLocal(event.GetGroup(), event.GetName())

		if gitopsObj == nil {
			return nil
		}

		gitopsWatcher := gitops.Shared.Watchers.Find(gitopsObj.GetGroupIdentifier())

		if gitopsWatcher != nil {
			if gitopsWatcher.Gitops.GetStatus().GetPending().Is(status.PENDING_SYNC, status.PENDING_DELETE) {
				return nil
			}

			gitopsObj.SetForceClone(true)
			gitopsObj.GetStatus().QueueState(status.CREATED)
			gitopsWatcher.GitopsQueue <- gitopsObj
		}
		break
	case events.EVENT_SYNC:
		gitopsObj := gitops.Shared.Registry.FindLocal(event.GetGroup(), event.GetName())

		if gitopsObj == nil {
			return nil
		}

		gitopsWatcher := gitops.Shared.Watchers.Find(gitopsObj.GetGroupIdentifier())

		if gitopsWatcher != nil {
			if gitopsWatcher.Gitops.GetStatus().GetPending().Is(status.PENDING_SYNC, status.PENDING_DELETE) {
				return nil
			}

			gitopsObj.SetForceSync(true)
			gitopsObj.GetStatus().QueueState(status.CLONING_GIT)
			gitopsWatcher.GitopsQueue <- gitopsObj
		}
		break
	case events.EVENT_INSPECT:
		gitopsObj := gitops.Shared.Registry.FindLocal(event.GetGroup(), event.GetName())

		if gitopsObj == nil {
			return nil
		}

		gitopsWatcher := gitops.Shared.Watchers.Find(gitopsObj.GetGroupIdentifier())

		if gitopsWatcher != nil {
			if gitopsWatcher.Gitops.GetStatus().GetPending().Is(status.PENDING_SYNC, status.PENDING_DELETE) {
				return nil
			}

			gitopsObj.GetStatus().QueueState(status.INSPECTING)
			gitopsWatcher.GitopsQueue <- gitopsObj
		}
		break
	}

	return nil

}

package reconcile

import (
	"context"
	"fmt"
	"github.com/simplecontainer/smr/implementations/gitops/gitops"
	"github.com/simplecontainer/smr/implementations/gitops/shared"
	"github.com/simplecontainer/smr/implementations/gitops/status"
	"github.com/simplecontainer/smr/implementations/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/manager"
	"go.uber.org/zap"
	"time"
)

func NewWatcher(gitopsObj *gitops.Gitops, mgr *manager.Manager, user *authentication.User) *watcher.Gitops {
	interval := 5 * time.Second
	ctx, fn := context.WithCancel(context.Background())

	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{fmt.Sprintf("/tmp/gitops.%s.%s.log", gitopsObj.Definition.Meta.Group, gitopsObj.Definition.Meta.Name)}

	loggerObj, err := cfg.Build()

	if err != nil {
		panic(err)
	}

	watcher := &watcher.Gitops{
		Gitops: gitopsObj,
		BackOff: watcher.BackOff{
			BackOff: false,
			Failure: 0,
		},
		Tracking:    true,
		Syncing:     false,
		GitopsQueue: make(chan *gitops.Gitops),
		User:        user,
		Ctx:         ctx,
		Cancel:      fn,
		Ticker:      time.NewTicker(interval),
		Logger:      loggerObj,
	}

	return watcher
}

func HandleTickerAndEvents(shared *shared.Shared, gitopsWatcher *watcher.Gitops) {
	for {
		select {
		case <-gitopsWatcher.Ctx.Done():
			gitopsWatcher.Ticker.Stop()
			close(gitopsWatcher.GitopsQueue)
			shared.Watcher.Remove(fmt.Sprintf("%s.%s", gitopsWatcher.Gitops.Definition.Meta.Group, gitopsWatcher.Gitops.Definition.Meta.Name))

			return
		case <-gitopsWatcher.GitopsQueue:
			gitopsWatcher.Ticker.Reset(5 * time.Second)
			go Gitops(shared, gitopsWatcher)
			break
		case <-gitopsWatcher.Ticker.C:
			if !gitopsWatcher.Gitops.Status.Reconciling && gitopsWatcher.Gitops.Status.GetCategory() != status.CATEGORY_END {
				go Gitops(shared, gitopsWatcher)
			} else {
				gitopsWatcher.Ticker.Stop()
			}

			break
		}
	}
}

func Gitops(shared *shared.Shared, gitopsWatcher *watcher.Gitops) {
	gitopsWatcher.Gitops.Status.Reconciling = true
	name := gitopsWatcher.Gitops.Definition.Meta.Name

	switch gitopsWatcher.Gitops.Status.GetState() {
	case status.STATUS_CREATED:
		gitopsWatcher.Logger.Info(fmt.Sprintf("%s is created", name))

		var err error
		gitopsWatcher.Gitops.Auth.AuthType, err = gitopsWatcher.Gitops.Prepare(shared.Client, gitopsWatcher.User)

		if err != nil {
			gitopsWatcher.Logger.Info(fmt.Sprintf("%s failed to solve gitops references", name))
			gitopsWatcher.Logger.Error(err.Error())
			gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INVALID_GIT)
			Loop(gitopsWatcher)
		}

		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONING_GIT)

		Loop(gitopsWatcher)
		break
	case status.STATUS_CLONING_GIT:
		auth, err := gitopsWatcher.Gitops.GetAuth()
		gitopsWatcher.Logger.Info(fmt.Sprintf("pulling latest changes from the repository: %s", gitopsWatcher.Gitops.RepoURL))

		if err != nil {
			gitopsWatcher.Logger.Info(fmt.Sprintf("%s failed to generate auth for the git repository", name))
			gitopsWatcher.Logger.Error(err.Error())
			gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INVALID_GIT)
			Loop(gitopsWatcher)

			return
		}

		err = gitopsWatcher.Gitops.CloneOrPull(auth)

		if err != nil {
			if err.Error() == gitops.POLLING_INTERVAL_ERROR {
				gitopsWatcher.Logger.Info("pulling latest changes will occur after polling interval")
			} else {
				gitopsWatcher.Logger.Info(fmt.Sprintf("%s failed to pull latest changes", name))
				gitopsWatcher.Logger.Error(err.Error())
				gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INVALID_GIT)
				Loop(gitopsWatcher)
			}
		} else {
			gitopsWatcher.Logger.Info(fmt.Sprintf("pulled latest changes from the repository: %s", gitopsWatcher.Gitops.RepoURL))
		}

		sync := false
		if !gitopsWatcher.Gitops.AutomaticSync {
			if gitopsWatcher.Gitops.ManualSync {
				sync = true
			} else {
				gitopsWatcher.Logger.Info("sync needs to be triggered manually")
			}
		} else {
			sync = true
		}

		if sync {
			if gitopsWatcher.Gitops.Status.PreviousState.State == status.STATUS_CREATED {
				gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_SYNCING)
			} else if gitopsWatcher.Gitops.Status.PreviousState.State == status.STATUS_INSYNC {
				gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INSPECTING)
			} else {
				gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_SYNCING)
			}

			Loop(gitopsWatcher)
		}

		break
	case status.STATUS_INVALID_GIT:
		gitopsWatcher.Logger.Info("git configuration is invalid or pull failed")

		break
	case status.STATUS_INVALID_DEFINITIONS:
		gitopsWatcher.Logger.Info("definitions are invalid")

		break
	case status.STATUS_SYNCING:
		gitopsWatcher.Logger.Info(fmt.Sprintf("attempt to sync commit %s", gitopsWatcher.Gitops.Commit.ID()))

		if gitopsWatcher.Syncing {
			gitopsWatcher.Logger.Info("gitops already reconciling, waiting for the free slot")
			return
		}

		gitopsWatcher.Syncing = true
		if gitopsWatcher.Gitops.Status.LastSyncedCommit != gitopsWatcher.Gitops.Commit.ID() || !gitopsWatcher.Gitops.Status.InSync {
			defs, err := gitopsWatcher.Gitops.Definitions(shared.Manager.RelationRegistry)

			err = gitopsWatcher.Gitops.Sync(shared.Client, gitopsWatcher.User, defs)

			if err != nil {
				if err.Error() == "object is same on the server" {
					gitopsWatcher.Logger.Info(fmt.Sprintf("gitops object is same on the server"))
				} else {
					gitopsWatcher.Logger.Info(fmt.Sprintf("failed to sync latest changes"))
					gitopsWatcher.Logger.Error(err.Error())
					gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INVALID_DEFINITIONS)
					gitopsWatcher.Syncing = false
					Loop(gitopsWatcher)
					return
				}
			}

			gitopsWatcher.Gitops.Status.LastSyncedCommit = gitopsWatcher.Gitops.Commit.ID()
			gitopsWatcher.Gitops.Status.InSync = true

			gitopsWatcher.Logger.Info(fmt.Sprintf("commit %s synced", gitopsWatcher.Gitops.Status.LastSyncedCommit))
		} else {
			gitopsWatcher.Logger.Info("everything synced")
		}

		gitopsWatcher.Syncing = false

		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INSYNC)
		Loop(gitopsWatcher)
		break
	case status.STATUS_INSPECTING:
		gitopsWatcher.Syncing = true
		if gitopsWatcher.Gitops.Status.LastSyncedCommit != gitopsWatcher.Gitops.Commit.ID() || !gitopsWatcher.Gitops.Status.InSync {
			defs, err := gitopsWatcher.Gitops.Definitions(shared.Manager.RelationRegistry)

			var drifted bool
			drifted, err = gitopsWatcher.Gitops.Drift(shared.Client, gitopsWatcher.User, defs)

			if err != nil {
				gitopsWatcher.Logger.Info(fmt.Sprintf("failed to compare latest changes"))
				gitopsWatcher.Logger.Error(err.Error())
				gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INVALID_GIT)
				Loop(gitopsWatcher)
				return
			}

			gitopsWatcher.Gitops.Status.InSync = drifted
			gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_DRIFTED)
		} else {
			gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INSYNC)
		}

		Loop(gitopsWatcher)
		break
	case status.STATUS_INSYNC:
		gitopsWatcher.Logger.Info("everything synced")
		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONING_GIT)

		// Let time ticker trigger transition state instead of direct transition
		gitopsWatcher.Gitops.Status.Reconciling = false
		break
	case status.STATUS_DRIFTED:
		gitopsWatcher.Logger.Info("drift detected")
		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_SYNCING)
		Loop(gitopsWatcher)
		break
	case status.STATUS_PENDING_DELETE:
		gitopsWatcher.Logger.Info("delete is in process")
		gitopsWatcher.Cancel()
		break
	}
}

func Loop(gitopsWatcher *watcher.Gitops) {
	gitopsWatcher.Gitops.Status.Reconciling = false
	gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
}

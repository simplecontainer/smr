package reconcile

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/status"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/static"
	"os"
	"strings"
	"time"
)

func NewWatcher(gitopsObj *implementation.Gitops, mgr *manager.Manager, user *authentication.User) *watcher.Gitops {
	interval := 5 * time.Second
	ctx, fn := context.WithCancel(context.Background())

	format := f.New(gitopsObj.Definition.GetPrefix(), "kind", static.KIND_GITOPS, gitopsObj.GetGroup(), gitopsObj.GetName())
	path := fmt.Sprintf("/tmp/%s", strings.Replace(format.ToString(), "/", "-", -1))

	loggerObj := logger.NewLogger(os.Getenv("LOG_LEVEL"), []string{path}, []string{path})

	gitopsObj.Status.Logger = loggerObj

	return &watcher.Gitops{
		Gitops:      gitopsObj,
		Children:    make([]*common.Request, 0),
		GitopsQueue: make(chan *implementation.Gitops),
		User:        user,
		Ctx:         ctx,
		Cancel:      fn,
		Ticker:      time.NewTicker(interval),
		Logger:      loggerObj,
		LogPath:     path,
	}
}

func HandleTickerAndEvents(shared *shared.Shared, gitopsWatcher *watcher.Gitops) {
	for {
		select {
		case <-gitopsWatcher.Ctx.Done():
			gitopsWatcher.Ticker.Stop()
			close(gitopsWatcher.GitopsQueue)

			for _, request := range gitopsWatcher.Children {
				go func() {
					err := request.AttemptRemove(shared.Manager.Http.Clients[shared.Manager.User.Username].Http, shared.Manager.Http.Clients[shared.Manager.User.Username].API)

					if err != nil {
						logger.Log.Error(err.Error())
					}
				}()
			}

			shared.Registry.Remove(gitopsWatcher.Gitops.Definition.GetPrefix(), gitopsWatcher.Gitops.Definition.Meta.Group, gitopsWatcher.Gitops.Definition.Meta.Name)
			shared.Watcher.Remove(fmt.Sprintf("%s.%s", gitopsWatcher.Gitops.Definition.Meta.Group, gitopsWatcher.Gitops.Definition.Meta.Name))
			logger.Log.Debug("gitops watcher deleted - proceed with deleting children")

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
	if gitopsWatcher.Gitops.Status.Reconciling {
		gitopsWatcher.Logger.Info("gitops already reconciling, waiting for the free slot")
		return
	}

	gitopsWatcher.Gitops.Status.Reconciling = true
	name := gitopsWatcher.Gitops.Definition.Meta.Name

	switch gitopsWatcher.Gitops.Status.GetState() {
	case status.STATUS_CREATED:
		shared.Registry.Sync(gitopsWatcher.Gitops)
		gitopsWatcher.Logger.Info(fmt.Sprintf("%s is created", name))

		err := gitopsWatcher.Gitops.Prepare(shared.Client, gitopsWatcher.User)

		if err != nil {
			gitopsWatcher.Logger.Info(fmt.Sprintf("%s failed to resolve gitops references and generate auth credentials", name))
			gitopsWatcher.Logger.Error(err.Error())
			gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INVALID_GIT)

			Loop(gitopsWatcher)
			return
		}

		gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONING_GIT)
		Loop(gitopsWatcher)
		break
	case status.STATUS_CLONING_GIT:
		shared.Registry.Sync(gitopsWatcher.Gitops)
		gitopsWatcher.Logger.Debug(fmt.Sprintf("attempting to clone/pull the: %s", gitopsWatcher.Gitops.RepoURL))

		err := gitopsWatcher.Gitops.Fetch(gitopsWatcher.LogPath)

		if err != nil {
			if errors.Is(err, implementation.ErrPoolingInterval) || errors.Is(err, git.NoErrAlreadyUpToDate) {
				gitopsWatcher.Logger.Debug(err.Error())
				gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONED_GIT)
			} else {
				gitopsWatcher.Logger.Info(fmt.Sprintf("%s failed to pull latest changes", name))
				gitopsWatcher.Logger.Error(err.Error())

				gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INVALID_GIT)
			}
		} else {
			gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONED_GIT)
		}

		Loop(gitopsWatcher)
		break
	case status.STATUS_CLONED_GIT:
		shared.Registry.Sync(gitopsWatcher.Gitops)
		gitopsWatcher.Logger.Debug(fmt.Sprintf("successfully cloned/pulled the: %s", gitopsWatcher.Gitops.RepoURL))

		if gitopsWatcher.ShouldSync() {
			gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_SYNCING)
			Loop(gitopsWatcher)
		} else {
			if !gitopsWatcher.Gitops.Status.LastSyncedCommit.IsZero() {
				gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INSPECTING)

				Loop(gitopsWatcher)
			}
		}

		logger.Log.Debug("reconcile is going to sleep till manual sync occured")
		gitopsWatcher.Gitops.Status.Reconciling = false
		gitopsWatcher.Ticker.Stop()
		break
	case status.STATUS_INVALID_GIT:
		shared.Registry.Sync(gitopsWatcher.Gitops)
		gitopsWatcher.Logger.Info("git configuration is invalid or pull failed")

		break
	case status.STATUS_INVALID_DEFINITIONS:
		shared.Registry.Sync(gitopsWatcher.Gitops)
		gitopsWatcher.Logger.Info("definitions are invalid")

		break
	case status.STATUS_SYNCING:
		shared.Registry.Sync(gitopsWatcher.Gitops)
		gitopsWatcher.Logger.Info(fmt.Sprintf("attempt to sync commit %s", gitopsWatcher.Gitops.Commit.ID()))

		if gitopsWatcher.Gitops.Status.LastSyncedCommit != gitopsWatcher.Gitops.Commit.ID() || !gitopsWatcher.Gitops.Status.InSync {
			defs, err := gitopsWatcher.Gitops.ReadDefinitions(shared.Manager.Kinds)

			if err != nil {
				gitopsWatcher.Logger.Error(err.Error())
				gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INVALID_DEFINITIONS)
			} else {
				if len(defs) == 0 {
					gitopsWatcher.Logger.Info(fmt.Sprintf("no valid definitions found: %s/%s", gitopsWatcher.Gitops.Path, gitopsWatcher.Gitops.DirectoryPath))
					gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INVALID_DEFINITIONS)
				} else {
					gitopsWatcher.Children, err = gitopsWatcher.Gitops.Sync(gitopsWatcher.Logger, shared.Client, gitopsWatcher.User, defs)

					if err != nil {
						gitopsWatcher.Logger.Info(fmt.Sprintf("failed to sync latest changes"))
						gitopsWatcher.Logger.Info(err.Error())

						gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INVALID_DEFINITIONS)

						Loop(gitopsWatcher)
						return
					}

					gitopsWatcher.Gitops.Status.LastSyncedCommit = gitopsWatcher.Gitops.Commit.ID()
					gitopsWatcher.Gitops.Status.InSync = true
					gitopsWatcher.Gitops.ManualSync = false

					gitopsWatcher.Logger.Info(fmt.Sprintf("commit %s synced", gitopsWatcher.Gitops.Status.LastSyncedCommit))
					gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INSYNC)
				}
			}
		} else {
			gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INSYNC)
			gitopsWatcher.Logger.Info("everything synced")
		}

		Loop(gitopsWatcher)
		break
	case status.STATUS_INSPECTING:
		shared.Registry.Sync(gitopsWatcher.Gitops)
		remoteHeadHash, err := gitopsWatcher.Gitops.RemoteHead()

		if err != nil {
			logger.Log.Error(err.Error())
			gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INVALID_GIT)
		} else {
			if gitopsWatcher.Gitops.Status.LastSyncedCommit != remoteHeadHash {
				gitopsWatcher.Gitops.ForcePoll = true
				gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONING_GIT)
			} else {
				defs, err := gitopsWatcher.Gitops.ReadDefinitions(shared.Manager.Kinds)

				var drifted bool
				drifted, err = gitopsWatcher.Gitops.Drift(shared.Client, gitopsWatcher.User, defs)

				if err != nil {
					gitopsWatcher.Logger.Info(fmt.Sprintf("failed to compare latest changes"))
					gitopsWatcher.Logger.Error(err.Error())
					gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INVALID_DEFINITIONS)

					Loop(gitopsWatcher)
					return
				}

				gitopsWatcher.Gitops.Status.InSync = drifted == false

				if gitopsWatcher.Gitops.Status.InSync {
					gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_INSYNC)
				} else {
					gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_DRIFTED)
				}
			}
		}

		Loop(gitopsWatcher)
		break
	case status.STATUS_INSYNC:
		shared.Registry.Sync(gitopsWatcher.Gitops)
		gitopsWatcher.Gitops.Status.Reconciling = false
		// NO-OP
		break
	case status.STATUS_DRIFTED:
		shared.Registry.Sync(gitopsWatcher.Gitops)
		if gitopsWatcher.Gitops.AutomaticSync {
			gitopsWatcher.Logger.Info("drift detected")
			gitopsWatcher.Gitops.Status.TransitionState(gitopsWatcher.Gitops.Definition.Meta.Name, status.STATUS_CLONED_GIT)

			Loop(gitopsWatcher)
		} else {
			gitopsWatcher.Gitops.Status.Reconciling = false
			gitopsWatcher.Ticker.Stop()
		}

		break
	case status.STATUS_PENDING_DELETE:
		shared.Registry.Sync(gitopsWatcher.Gitops)
		gitopsWatcher.Logger.Info("delete is in process")
		gitopsWatcher.Cancel()
		break
	}
}

func Loop(gitopsWatcher *watcher.Gitops) {
	gitopsWatcher.Gitops.Status.Reconciling = false
	gitopsWatcher.GitopsQueue <- gitopsWatcher.Gitops
}

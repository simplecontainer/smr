package reconcile

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/simplecontainer/smr/implementations/gitops/gitops"
	"github.com/simplecontainer/smr/implementations/gitops/shared"
	"github.com/simplecontainer/smr/implementations/gitops/watcher"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/definitions"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/manager"
	"go.uber.org/zap"
	"io"
	"net/http"
	"path"
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
			if gitopsWatcher.BackOff.BackOff {
				gitopsWatcher.Logger.Info("gitops reconcile is invalid, delete old and apply new",
					zap.String("repository", gitopsWatcher.Gitops.RepoURL),
				)
			} else {
				go SyncGitops(shared, gitopsWatcher)
			}
			break
		case <-gitopsWatcher.Ticker.C:
			if gitopsWatcher.BackOff.BackOff {
				gitopsWatcher.Logger.Info("gitops reconcile is invalid, delete old and apply new",
					zap.String("repository", gitopsWatcher.Gitops.RepoURL),
				)
			}

			if !gitopsWatcher.Gitops.AutomaticSync {
				if gitopsWatcher.Gitops.LastSyncedCommit.IsZero() {
					// For manual sync:
					// CheckInSync only after first successful sync
					break
				}
			}

			CheckInSync(shared, gitopsWatcher)

			if gitopsWatcher.Gitops.AutomaticSync && !gitopsWatcher.Gitops.InSync {
				go SyncGitops(shared, gitopsWatcher)
			}

			break
		}
	}
}

func ReconcileGitops(shared *shared.Shared, gitopsWatcher *watcher.Gitops) {}

func SyncGitops(shared *shared.Shared, gitopsWatcher *watcher.Gitops) {
	var orderedByDependencies []map[string]string
	var auth transport.AuthMethod
	var hash plumbing.Hash
	var err error

	if gitopsWatcher.Syncing {
		gitopsWatcher.Logger.Info("gitops already reconciling, waiting for the free slot")
		return
	}

	gitopsWatcher.Syncing = true

	gitopsWatcher.Gitops.AuthType, err = gitopsWatcher.Gitops.Prepare(shared.Client, gitopsWatcher.User)
	auth, err = GetAuth(gitopsWatcher.Gitops)

	if err != nil {
		gitopsWatcher.Logger.Error("gitops reconcile auth configuration failed",
			zap.String("repository", gitopsWatcher.Gitops.RepoURL),
			zap.Error(err),
		)

		gitopsWatcher.Backoff()
		gitopsWatcher.Syncing = false
		return
	}

	localPath := fmt.Sprintf("/tmp/%s", path.Base(gitopsWatcher.Gitops.RepoURL))
	hash, err = Clone(gitopsWatcher.Gitops, auth, localPath)

	if err != nil {
		gitopsWatcher.Logger.Error("gitops reconcile git clone failed",
			zap.String("repository", gitopsWatcher.Gitops.RepoURL),
			zap.Error(err),
		)

		gitopsWatcher.Backoff()
		gitopsWatcher.Syncing = false
		return
	}

	if gitopsWatcher.Gitops.LastSyncedCommit != hash || !gitopsWatcher.Gitops.InSync {
		orderedByDependencies, err = SortFiles(gitopsWatcher.Gitops, localPath, shared)

		if err != nil {
			gitopsWatcher.Logger.Error("gitops reconcile dependency order failed",
				zap.String("repository", gitopsWatcher.Gitops.RepoURL),
				zap.Error(err),
			)

			gitopsWatcher.Backoff()
			gitopsWatcher.Syncing = false
			return
		}

		if len(orderedByDependencies) > 0 {
			for _, fileInfo := range orderedByDependencies {
				fileName := fileInfo["name"]

				definition := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", localPath, gitopsWatcher.Gitops.DirectoryPath, fileName))
				if err != nil {
					gitopsWatcher.Logger.Info("gitops reconcile unable to read file",
						zap.String("repository", gitopsWatcher.Gitops.RepoURL),
						zap.Error(err),
					)

					gitopsWatcher.Backoff()
					gitopsWatcher.Syncing = false
					return
				}

				response := sendRequest(shared.Client, gitopsWatcher.User, "https://localhost:1443/api/v1/apply", definition, gitopsWatcher)

				if response.Success {
					gitopsWatcher.Logger.Info("gitops reconcile apply success",
						zap.String("repository", gitopsWatcher.Gitops.RepoURL),
						zap.String("file", fileName),
						zap.Error(err),
					)

					gitopsWatcher.Gitops.LastSyncedCommit = hash
				} else {
					gitopsWatcher.Logger.Info("gitops reconcile invalid object sent",
						zap.String("repository", gitopsWatcher.Gitops.RepoURL),
						zap.Error(err),
					)

					gitopsWatcher.Backoff()
					gitopsWatcher.Syncing = false
					return
				}
			}
		} else {
			gitopsWatcher.Logger.Info("git repository doesnt contain any valid simplecontainer definitions",
				zap.String("repository", gitopsWatcher.Gitops.RepoURL),
				zap.String("directory", gitopsWatcher.Gitops.DirectoryPath),
				zap.String("revision", gitopsWatcher.Gitops.Revision),
			)
		}
	} else {
		gitopsWatcher.Logger.Info("gitops reconcile hash is same as the last synced",
			zap.String("repository", gitopsWatcher.Gitops.RepoURL),
		)
	}

	gitopsWatcher.Syncing = false
}

func CheckInSync(shared *shared.Shared, gitopsWatcher *watcher.Gitops) {
	var auth transport.AuthMethod
	var err error

	if gitopsWatcher.Syncing {
		gitopsWatcher.Logger.Info("container already reconciling, waiting for the free slot")
		return
	}

	gitopsWatcher.Gitops.AuthType, err = gitopsWatcher.Gitops.Prepare(shared.Client, shared.Manager.User)
	auth, err = GetAuth(gitopsWatcher.Gitops)

	if err != nil {
		gitopsWatcher.Logger.Info("gitops reconcile auth configuration failed",
			zap.String("repository", gitopsWatcher.Gitops.RepoURL),
			zap.Error(err),
		)

		gitopsWatcher.Backoff()
		return
	}

	localPath := fmt.Sprintf("/tmp/%s", path.Base(gitopsWatcher.Gitops.RepoURL))
	_, err = Clone(gitopsWatcher.Gitops, auth, localPath)

	if err != nil {
		gitopsWatcher.Logger.Info("gitops reconcile git clone failed",
			zap.String("repository", gitopsWatcher.Gitops.RepoURL),
			zap.Error(err),
		)

		gitopsWatcher.Backoff()
	}

	gitopsWatcher.Logger.Info("checking if everything is in sync",
		zap.String("repoUrl", gitopsWatcher.Gitops.RepoURL),
	)

	orderedByDependencies, err := SortFiles(gitopsWatcher.Gitops, localPath, shared)

	if err != nil {
		gitopsWatcher.Logger.Error("gitops reconcile git clone failed",
			zap.String("repository", gitopsWatcher.Gitops.RepoURL),
			zap.Error(err),
		)

		gitopsWatcher.Backoff()
	}

	inSync := true

	for _, fileInfo := range orderedByDependencies {
		fileName := fileInfo["name"]

		gitopsWatcher.Logger.Info("checking in sync", zap.String("file", fileName))

		definition := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", localPath, gitopsWatcher.Gitops.DirectoryPath, fileName))
		if err != nil {
			gitopsWatcher.Logger.Debug("gitops reconcile unable to read file",
				zap.String("repository", gitopsWatcher.Gitops.RepoURL),
				zap.Error(err),
			)

			gitopsWatcher.Backoff()
		}

		response := sendRequest(shared.Client, gitopsWatcher.User, "https://localhost:1443/api/v1/compare", definition, gitopsWatcher)

		switch response.HttpStatus {
		case http.StatusOK:
			gitopsWatcher.Logger.Info("file is in sync", zap.String("file", fileName))
			break
		case http.StatusTeapot:
			gitopsWatcher.Logger.Info("file is drifted", zap.String("file", fileName))
			inSync = false
			break
		case http.StatusBadRequest:
			gitopsWatcher.Logger.Info("gitops reconcile invalid object sent",
				zap.String("repository", gitopsWatcher.Gitops.RepoURL),
				zap.Error(err),
			)

			gitopsWatcher.Backoff()
			break
		}
	}

	gitopsWatcher.Gitops.InSync = inSync
}

func sendRequest(client *client.Http, user *authentication.User, URL string, data string, gitopsWatcher *watcher.Gitops) *httpcontract.ResponseImplementation {
	var req *http.Request
	var err error

	if len(data) > 0 {
		if err != nil {
			return &httpcontract.ResponseImplementation{
				HttpStatus:       0,
				Explanation:      "failed to marshal data for sending request",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
			}
		}

		req, err = http.NewRequest("POST", URL, bytes.NewBuffer([]byte(data)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Owner", fmt.Sprintf("gitops.%s.%s", gitopsWatcher.Gitops.Definition.Meta.Group, gitopsWatcher.Gitops.Definition.Meta.Name))
	} else {
		req, err = http.NewRequest("GET", URL, nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Owner", fmt.Sprintf("gitops.%s.%s", gitopsWatcher.Gitops.Definition.Meta.Group, gitopsWatcher.Gitops.Definition.Meta.Name))
	}

	if err != nil {
		return &httpcontract.ResponseImplementation{
			HttpStatus:       0,
			Explanation:      "failed to craft request",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}

	resp, err := client.Get(user.Username).Http.Do(req)

	if err != nil {
		return &httpcontract.ResponseImplementation{
			HttpStatus:       0,
			Explanation:      "failed to connect to the smr-agent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &httpcontract.ResponseImplementation{
			HttpStatus:       0,
			Explanation:      "invalid response from the smr-agent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}

	var response httpcontract.ResponseImplementation
	err = json.Unmarshal(body, &response)

	if err != nil {
		return &httpcontract.ResponseImplementation{
			HttpStatus:       0,
			Explanation:      "failed to unmarshal body response from smr-agent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}

	return &response
}

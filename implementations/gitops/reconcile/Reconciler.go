package reconcile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/qdnqn/smr/implementations/gitops/gitops"
	"github.com/qdnqn/smr/implementations/gitops/shared"
	"github.com/qdnqn/smr/implementations/gitops/watcher"
	"github.com/qdnqn/smr/pkg/definitions"
	v1 "github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/httpcontract"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/qdnqn/smr/pkg/manager"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

func NewWatcher(gitopsObj *v1.Gitops) *watcher.Gitops {
	interval := 5 * time.Second
	ctx, fn := context.WithCancel(context.Background())

	watcher := &watcher.Gitops{
		Gitops: &gitops.Gitops{
			RepoURL:          gitopsObj.Spec.RepoURL,
			Revision:         gitopsObj.Spec.Revision,
			DirectoryPath:    gitopsObj.Spec.DirectoryPath,
			PoolingInterval:  gitopsObj.Spec.PoolingInterval,
			AutomaticSync:    gitopsObj.Spec.AutomaticSync,
			InSync:           false,
			CertKeyRef:       gitopsObj.Spec.CertKeyRef,
			HttpAuthRef:      gitopsObj.Spec.HttpAuthRef,
			LastSyncedCommit: plumbing.Hash{},
			CertKey:          nil,
			HttpAuth:         nil,
			Definition:       gitopsObj,
		},
		GitopsQueue: make(chan *gitops.Gitops),
		Ctx:         ctx,
		Cancel:      fn,
		Ticker:      time.NewTicker(interval),
	}

	return watcher
}

func HandleTickerAndEvents(shared *shared.Shared, gitopsWatcher *watcher.Gitops) {
	for {
		select {
		case <-gitopsWatcher.Ctx.Done():
			gitopsWatcher.Ticker.Stop()
			close(gitopsWatcher.GitopsQueue)
			shared.Watcher.Remove(fmt.Sprintf("%s.%s", gitopsWatcher.Gitops.Definition.Meta.Group, gitopsWatcher.Gitops.Definition.Meta.Identifier))

			return
		case <-gitopsWatcher.GitopsQueue:
			go ReconcileGitops(shared, gitopsWatcher)
			break
		case <-gitopsWatcher.Ticker.C:
			if gitopsWatcher.Gitops.AutomaticSync {
				go ReconcileGitops(shared, gitopsWatcher)
			}
			break
		}
	}
}

func ReconcileGitops(shared *shared.Shared, gitopsWatcher *watcher.Gitops) {
	gitopsObj := gitopsWatcher.Gitops

	if gitopsWatcher.Syncing {
		logger.Log.Info("container already reconciling, waiting for the free slot")
		return
	}

	var auth transport.AuthMethod = nil

	if gitopsObj.HttpAuth != nil {
		auth = &gitHttp.BasicAuth{
			Username: gitopsObj.HttpAuth.Username,
			Password: gitopsObj.HttpAuth.Password,
		}
	}

	if gitopsObj.CertKey != nil {
		auth, _ = ssh.NewPublicKeys(ssh.DefaultUsername, []byte(gitopsObj.CertKey.PrivateKey), gitopsObj.CertKey.PrivateKeyPassword)
	}

	localPath := fmt.Sprintf("/tmp/%s", path.Base(gitopsObj.RepoURL))

	if _, err := os.Stat(localPath); errors.Is(err, os.ErrNotExist) {
		_, err := git.PlainClone(localPath, false, &git.CloneOptions{
			URL:      gitopsObj.RepoURL,
			Progress: os.Stdout,
			Auth:     auth,
		})

		if err != nil {
			logger.Log.Error("failed to fetch repository", zap.String("repository", gitopsObj.RepoURL))
		}
	}

	r, _ := git.PlainOpen(localPath)

	w, _ := r.Worktree()

	_ = w.Pull(&git.PullOptions{RemoteName: "origin"})

	ref, _ := r.Head()
	r.CommitObject(ref.Hash())

	logger.Log.Debug("pulled the latest changes from the git repository", zap.String("repoUrl", gitopsObj.RepoURL))

	if gitopsObj.LastSyncedCommit != ref.Hash() {
		entries, err := os.ReadDir(fmt.Sprintf("%s/%s", localPath, gitopsObj.DirectoryPath))
		if err != nil {
			logger.Log.Error(err.Error())
		}

		orderedByDependencies := make([]map[string]string, 0)

		for _, e := range entries {
			definition := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", localPath, gitopsObj.DirectoryPath, e.Name()))
			if err != nil {
				log.Fatalf("unable to read file: %v", err)
			}

			data := make(map[string]interface{})

			err = json.Unmarshal([]byte(definition), &data)
			if err != nil {
				logger.Log.Error("invalid json defined for the object", zap.String("error", err.Error()))
			}

			position := -1

			for index, orderedEntry := range orderedByDependencies {
				deps := shared.Manager.DefinitionRegistry.GetDependencies(orderedEntry["kind"])

				for _, dp := range deps {
					if data["kind"].(string) == dp {
						position = index
					}
				}
			}

			if position != -1 {
				orderedByDependencies = append(orderedByDependencies[:position+1], orderedByDependencies[position:]...)
				orderedByDependencies[position] = map[string]string{"name": e.Name(), "kind": data["kind"].(string)}
			} else {
				orderedByDependencies = append(orderedByDependencies, map[string]string{"name": e.Name(), "kind": data["kind"].(string)})
			}
		}

		for _, fileInfo := range orderedByDependencies {
			fileName := fileInfo["name"]

			logger.Log.Debug("trying to reconcile", zap.String("file", fileName))

			definition := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", localPath, gitopsObj.DirectoryPath, fileName))
			if err != nil {
				log.Fatalf("unable to read file: %v", err)
			}

			client, err := shared.Manager.Keys.GenerateHttpClient()

			if err != nil {
				logger.Log.Error("gitops reconciler failed to generate http client for the mtls")
			}

			response := sendRequest(client, "https://localhost:1443/api/v1/apply", definition, gitopsObj)

			if response.Success {
				logger.Log.Debug("gitops response collected", zap.String("response", response.Explanation))
			} else {
				logger.Log.Debug("gitops response collected", zap.String("response", response.Explanation), zap.String("error", response.ErrorExplanation))
			}
		}

		gitopsObj.LastSyncedCommit = ref.Hash()
	} else {
		logger.Log.Debug("checking if everything is in sync", zap.String("repoUrl", gitopsObj.RepoURL))
	}
}

func CheckInSync(mgr *manager.Manager, gitopsWatcher *watcher.Gitops) {
	gitopsObj := gitopsWatcher.Gitops
	var auth transport.AuthMethod = nil

	if gitopsObj.HttpAuth != nil {
		auth = &gitHttp.BasicAuth{
			Username: gitopsObj.HttpAuth.Username,
			Password: gitopsObj.HttpAuth.Password,
		}
	}

	if gitopsObj.CertKey != nil {
		auth, _ = ssh.NewPublicKeys(ssh.DefaultUsername, []byte(gitopsObj.CertKey.PrivateKey), gitopsObj.CertKey.PrivateKeyPassword)
	}

	localPath := fmt.Sprintf("/tmp/%s", path.Base(gitopsObj.RepoURL))

	if _, err := os.Stat(localPath); errors.Is(err, os.ErrNotExist) {
		_, err := git.PlainClone(localPath, false, &git.CloneOptions{
			URL:      gitopsObj.RepoURL,
			Progress: os.Stdout,
			Auth:     auth,
		})

		if err != nil {
			logger.Log.Error("failed to fetch repository", zap.String("repository", gitopsObj.RepoURL))
		}
	}

	r, _ := git.PlainOpen(localPath)

	w, _ := r.Worktree()

	_ = w.Pull(&git.PullOptions{RemoteName: "origin"})

	ref, _ := r.Head()
	r.CommitObject(ref.Hash())

	logger.Log.Debug("pulled the latest changes from the git repository", zap.String("repoUrl", gitopsObj.RepoURL))
	logger.Log.Debug("checking if everything is in sync", zap.String("repoUrl", gitopsObj.RepoURL))

	entries, err := os.ReadDir(fmt.Sprintf("%s/%s", localPath, gitopsObj.DirectoryPath))
	if err != nil {
		logger.Log.Error(err.Error())
	}

	orderedByDependencies := make([]map[string]string, 0)

	for _, e := range entries {
		definition := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", localPath, gitopsObj.DirectoryPath, e.Name()))
		if err != nil {
			log.Fatalf("unable to read file: %v", err)
		}

		data := make(map[string]interface{})

		err := json.Unmarshal([]byte(definition), &data)
		if err != nil {
			logger.Log.Error("invalid json defined for the object", zap.String("error", err.Error()))
		}

		position := -1

		for index, orderedEntry := range orderedByDependencies {
			deps := mgr.DefinitionRegistry.GetDependencies(orderedEntry["kind"])

			for _, dp := range deps {
				if data["kind"].(string) == dp {
					position = index
				}
			}
		}

		if position != -1 {
			orderedByDependencies = append(orderedByDependencies[:position+1], orderedByDependencies[position:]...)
			orderedByDependencies[position] = map[string]string{"name": e.Name(), "kind": data["kind"].(string)}
		} else {
			orderedByDependencies = append(orderedByDependencies, map[string]string{"name": e.Name(), "kind": data["kind"].(string)})
		}
	}

	inSync := true

	for _, fileInfo := range orderedByDependencies {
		fileName := fileInfo["name"]

		logger.Log.Debug("checking in sync", zap.String("file", fileName))

		definition := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", localPath, gitopsObj.DirectoryPath, fileName))
		if err != nil {
			log.Fatalf("unable to read file: %v", err)
		}

		client, err := mgr.Keys.GenerateHttpClient()

		if err != nil {
			logger.Log.Error("gitops reconciler failed to generate http client for the mtls")
		}

		response := sendRequest(client, "https://localhost:1443/api/v1/compare", definition, gitopsObj)

		switch response.HttpStatus {
		case http.StatusOK:
			logger.Log.Debug("file is in sync", zap.String("file", fileName))
			break
		case http.StatusTeapot:
			logger.Log.Debug("file is drifted", zap.String("file", fileName))
			inSync = false
			break
		case http.StatusBadRequest:
			logger.Log.Error(response.ErrorExplanation)
			break
		}
	}

	gitopsObj.InSync = inSync
}

func sendRequest(client *http.Client, URL string, data string, gitopsObj *gitops.Gitops) *httpcontract.ResponseImplementation {
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
		req.Header.Set("Owner", fmt.Sprintf("gitops.%s.%s", gitopsObj.Definition.Meta.Group, gitopsObj.Definition.Meta.Identifier))
	} else {
		req, err = http.NewRequest("GET", URL, nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Owner", fmt.Sprintf("gitops.%s.%s", gitopsObj.Definition.Meta.Group, gitopsObj.Definition.Meta.Identifier))
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

	resp, err := client.Do(req)

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

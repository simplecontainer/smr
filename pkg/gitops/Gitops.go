package gitops

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
	"github.com/qdnqn/smr/pkg/definitions"
	"github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/httpcontract"
	"github.com/qdnqn/smr/pkg/keys"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/qdnqn/smr/pkg/objectdependency"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

func NewWatcher(gitops v1.Gitops) *Gitops {
	if gitops.Spec.PoolingInterval == "" {
		gitops.Spec.PoolingInterval = "30s"
	}

	interval, err := time.ParseDuration(gitops.Spec.PoolingInterval)

	if err != nil {
		logger.Log.Error(err.Error())
		return nil
	}

	return &Gitops{
		RepoURL:         gitops.Spec.RepoURL,
		Revision:        gitops.Spec.Revision,
		DirectoryPath:   gitops.Spec.DirectoryPath,
		PoolingInterval: gitops.Spec.PoolingInterval,
		AutomaticSync:   gitops.Spec.AutomaticSync,
		CertKeyRef:      gitops.Spec.CertKeyRef,
		HttpAuthRef:     gitops.Spec.HttpAuthRef,
		Definition:      gitops,
		CertKey:         nil,
		HttpAuth:        nil,
		GitopsQueue:     make(chan Event),
		Ctx:             context.Background(),
		Ticker:          time.NewTicker(interval),
	}
}

func (gitops *Gitops) HandleTickerAndEvents(definitionRegistry *objectdependency.DefinitionRegistry, keys *keys.Keys) {
	for {
		select {
		case <-gitops.Ctx.Done():
			return
			break
		case event := <-gitops.GitopsQueue:
			gitops.HandleEvent(event)
			break
		case t := <-gitops.Ticker.C:
			gitops.CheckInSync(definitionRegistry, keys)

			if gitops.AutomaticSync {
				logger.Log.Debug("triggering gitops auto sync", zap.String("ticker", t.String()))
				gitops.ReconcileGitOps(definitionRegistry, keys)
			}
			break
		}
	}
}

func (gitops *Gitops) HandleEvent(event Event) {
	switch event.Event {
	case RESTART:
		gitops.LastSyncedCommit = plumbing.Hash{0}
		break
	case STOP:
		gitops.Ticker.Stop()
		break
	case KILL:
		gitops.Ticker.Stop()
		gitops.Ctx.Done()
		close(gitops.GitopsQueue)
	}
}

func (gitops *Gitops) ReconcileGitOps(definitionRegistry *objectdependency.DefinitionRegistry, keys *keys.Keys) {
	var auth transport.AuthMethod = nil

	if gitops.HttpAuth != nil {
		auth = &gitHttp.BasicAuth{
			Username: gitops.HttpAuth.Username,
			Password: gitops.HttpAuth.Password,
		}
	}

	if gitops.CertKey != nil {
		auth, _ = ssh.NewPublicKeys(ssh.DefaultUsername, []byte(gitops.CertKey.PrivateKey), gitops.CertKey.PrivateKeyPassword)
	}

	localPath := fmt.Sprintf("/tmp/%s", path.Base(gitops.RepoURL))

	if _, err := os.Stat(localPath); errors.Is(err, os.ErrNotExist) {
		_, err := git.PlainClone(localPath, false, &git.CloneOptions{
			URL:      gitops.RepoURL,
			Progress: os.Stdout,
			Auth:     auth,
		})

		if err != nil {
			logger.Log.Error("failed to fetch repository", zap.String("repository", gitops.RepoURL))
		}
	}

	r, _ := git.PlainOpen(localPath)

	w, _ := r.Worktree()

	_ = w.Pull(&git.PullOptions{RemoteName: "origin"})

	ref, _ := r.Head()
	r.CommitObject(ref.Hash())

	logger.Log.Debug("pulled the latest changes from the git repository", zap.String("repoUrl", gitops.RepoURL))

	if gitops.LastSyncedCommit != ref.Hash() {
		entries, err := os.ReadDir(fmt.Sprintf("%s/%s", localPath, gitops.DirectoryPath))
		if err != nil {
			logger.Log.Error(err.Error())
		}

		orderedByDependencies := make([]map[string]string, 0)

		for _, e := range entries {
			definition := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", localPath, gitops.DirectoryPath, e.Name()))
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
				deps := definitionRegistry.GetDependencies(orderedEntry["kind"])

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

			definition := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", localPath, gitops.DirectoryPath, fileName))
			if err != nil {
				log.Fatalf("unable to read file: %v", err)
			}

			client, err := keys.GenerateHttpClient()

			if err != nil {
				logger.Log.Error("gitops reconciler failed to generate http client for the mtls")
			}

			response := gitops.sendRequest(client, "https://localhost:1443/api/v1/apply", definition)

			if response.Success {
				logger.Log.Debug("gitops response collected", zap.String("response", response.Explanation))
			} else {
				logger.Log.Debug("gitops response collected", zap.String("response", response.Explanation), zap.String("error", response.ErrorExplanation))
			}
		}

		gitops.LastSyncedCommit = ref.Hash()
	} else {
		logger.Log.Debug("checking if everything is in sync", zap.String("repoUrl", gitops.RepoURL))
	}
}

func (gitops *Gitops) CheckInSync(definitionRegistry *objectdependency.DefinitionRegistry, keys *keys.Keys) {
	var auth transport.AuthMethod = nil

	if gitops.HttpAuth != nil {
		auth = &gitHttp.BasicAuth{
			Username: gitops.HttpAuth.Username,
			Password: gitops.HttpAuth.Password,
		}
	}

	if gitops.CertKey != nil {
		auth, _ = ssh.NewPublicKeys(ssh.DefaultUsername, []byte(gitops.CertKey.PrivateKey), gitops.CertKey.PrivateKeyPassword)
	}

	localPath := fmt.Sprintf("/tmp/%s", path.Base(gitops.RepoURL))

	if _, err := os.Stat(localPath); errors.Is(err, os.ErrNotExist) {
		_, err := git.PlainClone(localPath, false, &git.CloneOptions{
			URL:      gitops.RepoURL,
			Progress: os.Stdout,
			Auth:     auth,
		})

		if err != nil {
			logger.Log.Error("failed to fetch repository", zap.String("repository", gitops.RepoURL))
		}
	}

	r, _ := git.PlainOpen(localPath)

	w, _ := r.Worktree()

	_ = w.Pull(&git.PullOptions{RemoteName: "origin"})

	ref, _ := r.Head()
	r.CommitObject(ref.Hash())

	logger.Log.Debug("pulled the latest changes from the git repository", zap.String("repoUrl", gitops.RepoURL))
	logger.Log.Debug("checking if everything is in sync", zap.String("repoUrl", gitops.RepoURL))

	entries, err := os.ReadDir(fmt.Sprintf("%s/%s", localPath, gitops.DirectoryPath))
	if err != nil {
		logger.Log.Error(err.Error())
	}

	orderedByDependencies := make([]map[string]string, 0)

	for _, e := range entries {
		definition := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", localPath, gitops.DirectoryPath, e.Name()))
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
			deps := definitionRegistry.GetDependencies(orderedEntry["kind"])

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

		definition := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", localPath, gitops.DirectoryPath, fileName))
		if err != nil {
			log.Fatalf("unable to read file: %v", err)
		}

		client, err := keys.GenerateHttpClient()

		if err != nil {
			logger.Log.Error("gitops reconciler failed to generate http client for the mtls")
		}

		response := gitops.sendRequest(client, "https://localhost:1443/api/v1/compare", definition)

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

	gitops.InSync = inSync
}

func (gitops *Gitops) sendRequest(client *http.Client, URL string, data string) *httpcontract.ResponseImplementation {
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
		req.Header.Set("Owner", fmt.Sprintf("gitops.%s.%s", gitops.Definition.Meta.Group, gitops.Definition.Meta.Identifier))
	} else {
		req, err = http.NewRequest("GET", URL, nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Owner", fmt.Sprintf("gitops.%s.%s", gitops.Definition.Meta.Group, gitops.Definition.Meta.Identifier))
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

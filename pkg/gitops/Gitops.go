package gitops

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/qdnqn/smr/pkg/definitions"
	"github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/dependency"
	"github.com/qdnqn/smr/pkg/httpcontract"
	"github.com/qdnqn/smr/pkg/keys"
	"github.com/qdnqn/smr/pkg/logger"
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

func (gitops *Gitops) HandleTickerAndEvents(definitionRegistry *dependency.DefinitionRegistry, keys *keys.Keys) {
	for {
		select {
		case <-gitops.Ctx.Done():
			return
			break
		case event := <-gitops.GitopsQueue:
			gitops.HandleEvent(event)
			break
		case t := <-gitops.Ticker.C:
			if !gitops.AutomaticSync {
				logger.Log.Info("triggering gitops sync is set to manual", zap.String("repository", gitops.RepoURL))
				gitops.Ticker.Stop()
			} else {
				logger.Log.Info("triggering gitops sync from the remote repository", zap.String("ticker", t.String()))
				gitops.ReconcileGitOps(definitionRegistry, keys)
			}
			break
		}
	}
}

func (gitops *Gitops) HandleEvent(event Event) {
	switch event.Event {
	case STOP:
		gitops.Ticker.Stop()
		break
	case KILL:
		gitops.Ticker.Stop()
		gitops.Ctx.Done()
		close(gitops.GitopsQueue)
	}
}

func (gitops *Gitops) ReconcileGitOps(definitionRegistry *dependency.DefinitionRegistry, keys *keys.Keys) {
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

	logger.Log.Info("pulled the latest changes from the git repository", zap.String("repoUrl", gitops.RepoURL))

	if gitops.LastSyncedCommit != ref.Hash() {
		entries, err := os.ReadDir(fmt.Sprintf("%s/%s", localPath, gitops.DirectoryPath))
		if err != nil {
			logger.Log.Error(err.Error())
		}

		var orderedByDependencies []string

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

			dependencies := definitionRegistry.GetDependencies(data["kind"].(string))

			if len(dependencies) == 0 {
				orderedByDependencies = append(orderedByDependencies, e.Name())
			} else {
				// TODO: Order by dependencies somehow?
				for _, oe := range orderedByDependencies {
					fmt.Println(oe)
				}
			}
		}

		for _, fileName := range orderedByDependencies {
			logger.Log.Info("trying to reconcile", zap.String("file", fileName))

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
				logger.Log.Info("gitops response collected", zap.String("response", response.Explanation))
			} else {
				logger.Log.Info("gitops response collected", zap.String("response", response.Explanation), zap.String("error", response.ErrorExplanation))
			}
		}

		gitops.LastSyncedCommit = ref.Hash()
	} else {
		logger.Log.Info("everything synced", zap.String("repoUrl", gitops.RepoURL))
	}
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
	} else {
		req, err = http.NewRequest("GET", URL, nil)
		req.Header.Set("Content-Type", "application/json")
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

package gitops

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/imroc/req/v3"
	"go.uber.org/zap"
	"log"
	"os"
	"path"
	"smr/pkg/definitions"
	"smr/pkg/logger"
	"time"
)

func NewWatcher(gitops definitions.Gitops) *Gitops {
	return &Gitops{
		RepoURL:         gitops.Spec.RepoURL,
		Revision:        gitops.Spec.Revision,
		DirectoryPath:   gitops.Spec.DirectoryPath,
		PoolingInterval: gitops.Spec.PoolingInterval,
		CertKeyRef:      gitops.Spec.CertKeyRef,
		HttpAuthRef:     gitops.Spec.HttpAuthRef,
		CertKey:         nil,
		HttpAuth:        nil,
		GitopsQueue:     make(chan Event),
		Ctx:             context.Background(),
		Ticker:          time.NewTicker(gitops.Spec.PoolingInterval),
	}
}

func (gitops *Gitops) HandleTickerAndEvents() {
	for {
		select {
		case <-gitops.Ctx.Done():
			return
			break
		case event := <-gitops.GitopsQueue:
			gitops.HandleEvent(event)
			break
		case t := <-gitops.Ticker.C:
			logger.Log.Info("triggering gitops sync from the remote repository", zap.String("ticker", t.String()))
			gitops.ReconcileGitOps()
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

func (gitops *Gitops) ReconcileGitOps() {
	var auth transport.AuthMethod = nil

	if gitops.HttpAuth != nil {
		auth = &http.BasicAuth{
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
	} else {
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

			for _, e := range entries {
				logger.Log.Info("trying to reconcile", zap.String("file", e.Name()))

				definition := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", localPath, gitops.DirectoryPath, e.Name()))
				if err != nil {
					log.Fatalf("unable to read file: %v", err)
				}

				gitops.sendRequest("http://localhost:8080/apply", definition)
			}

			gitops.LastSyncedCommit = ref.Hash()
		} else {
			logger.Log.Info("everything synced", zap.String("repoUrl", gitops.RepoURL))
		}
	}
}

// TODO:Refactor later

type Result struct {
	Data string `json:"data"`
}

func (gitops *Gitops) sendRequest(URL string, jsonData string) {
	client := req.C().DevMode()
	var result Result

	resp, err := client.R().
		SetBodyJsonString(jsonData).
		SetSuccessResult(&result).
		Post(URL)

	if err != nil {
		logger.Log.Error(err.Error())
	}

	if !resp.IsSuccessState() {
		fmt.Println("bad response status:", resp.Status)
		return
	}
}

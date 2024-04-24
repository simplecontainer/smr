package gitops

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"os"
	"path"
	"smr/pkg/definitions"
	"smr/pkg/logger"
	"time"
  "github.com/go-git/go-git/v5"
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
	}
}

func (gitops *Gitops) RunWatcher() {
	go gitops.GitopsServer()
}

func (gitops *Gitops) GitopsServer() {
	for {
		localPath := fmt.Sprintf("/tmp/%s", path.Base(gitops.RepoURL))

		if _, err := os.Stat(localPath); errors.Is(err, os.ErrNotExist) {
			_, err := git.PlainClone(localPath, false, &git.CloneOptions{
				URL:      gitops.RepoURL,
				Progress: os.Stdout,
			})

			if err != nil {
				logger.Log.Error("failed to fetch repository", zap.String("repository", gitops.RepoURL))
			}
		} else {
			r, _ := git.PlainOpen(localPath)

			w, _ := r.Worktree()

			_ = w.Pull(&git.PullOptions{RemoteName: "origin"})

			ref, _ := r.Head()
			commit, _ := r.CommitObject(ref.Hash())

			fmt.Println(commit.String())
		}

		time.Sleep(60 * time.Second)
	}
}
package gitops

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"os"
	"time"
)

const POLLING_INTERVAL_ERROR = "latest changes not pulled - waiting for interval"

func (gitops *Gitops) CloneOrPull(auth transport.AuthMethod) error {
	var repository *git.Repository
	var err error

	if _, err = os.Stat(gitops.Path); errors.Is(err, os.ErrNotExist) {
		repository, err = git.PlainClone(gitops.Path, false, &git.CloneOptions{
			URL:           gitops.RepoURL,
			Progress:      os.Stdout,
			ReferenceName: plumbing.NewBranchReferenceName(gitops.Revision),
			Auth:          auth,
		})

		if err != nil {
			return err
		}
	} else {
		repository, err = git.PlainOpen(gitops.Path)

		if err != nil {
			return err
		}
	}

	worktree, _ := repository.Worktree()

	var d time.Duration
	d, err = time.ParseDuration(gitops.PoolingInterval)

	if err != nil {
		d = time.Second * 180
	}

	var ref *plumbing.Reference

	if time.Now().Sub(gitops.LastPoll) > d || gitops.ForcePoll {
		err = worktree.Pull(&git.PullOptions{
			RemoteName:    "origin",
			Auth:          auth,
			SingleBranch:  true,
			Force:         true,
			ReferenceName: plumbing.NewBranchReferenceName(gitops.Revision),
		})

		if err != nil {
			if err.Error() != "already up-to-date" {
				return err
			}
		}

		gitops.LastPoll = time.Now()
		gitops.ForcePoll = false

		ref, _ = repository.Head()
		gitops.Commit, err = repository.CommitObject(ref.Hash())

		if err != nil {
			return err
		}

		if _, missing := os.Stat(fmt.Sprintf("%s/%s", gitops.Path, gitops.DirectoryPath)); os.IsNotExist(missing) {
			return errors.New(fmt.Sprintf("%s does not exists in the repository"))
		}

		return nil
	} else {
		return errors.New(POLLING_INTERVAL_ERROR)
	}
}

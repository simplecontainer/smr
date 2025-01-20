package implementation

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"os"
	"time"
)

var (
	ErrPoolingInterval = errors.New("sleep interval didn't pass")
)

func (gitops *Gitops) Fetch(logpath string) error {
	if _, err := git.PlainOpen(gitops.Path); err != nil {
		err = gitops.Clone(gitops.AuthResolved, logpath)

		fmt.Println(err)

		if err != nil {
			if errors.Is(err, git.ErrRepositoryAlreadyExists) {
				err = gitops.Pull(gitops.AuthResolved, logpath)

				if err != nil {
					if errors.Is(err, git.NoErrAlreadyUpToDate) {
						fmt.Println(err)
						return nil
					}

					fmt.Println(err)

					return err
				}
			}

			return err
		}

		return gitops.Pull(gitops.AuthResolved, logpath)
	} else {
		var sleepDuration time.Duration
		sleepDuration, err = time.ParseDuration(gitops.PoolingInterval)

		if err != nil {
			return err
		}

		if time.Now().Sub(gitops.LastPoll) > sleepDuration || gitops.ForcePoll {
			return gitops.Pull(gitops.AuthResolved, logpath)
		} else {
			return ErrPoolingInterval
		}
	}
}

func (gitops *Gitops) Clone(auth transport.AuthMethod, logpath string) error {
	file, err := os.OpenFile(logpath, os.O_RDWR, 0644)
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	if err != nil {
		return err
	}

	_, err = git.PlainClone(gitops.Path, false, &git.CloneOptions{
		URL:           gitops.RepoURL,
		Progress:      file,
		ReferenceName: plumbing.NewBranchReferenceName(gitops.Revision),
		Auth:          auth,
	})

	if err != nil {
		return err
	}

	return nil
}

func (gitops *Gitops) Pull(auth transport.AuthMethod, logpath string) error {
	file, err := os.OpenFile(logpath, os.O_RDWR, 0644)
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	if err != nil {
		return err
	}

	repository, err := git.PlainOpen(gitops.Path)

	if err != nil {
		return err
	}

	worktree, _ := repository.Worktree()

	var ref *plumbing.Reference

	err = worktree.Pull(&git.PullOptions{
		RemoteName:    "origin",
		Auth:          auth,
		SingleBranch:  true,
		Force:         true,
		Progress:      file,
		ReferenceName: plumbing.NewBranchReferenceName(gitops.Revision),
	})

	gitops.LastPoll = time.Now()
	gitops.ForcePoll = false

	ref, err = repository.Head()

	if err != nil {
		return err
	}

	gitops.Commit, err = repository.CommitObject(ref.Hash())
	return err
}

func (gitops *Gitops) RemoteHead() (plumbing.Hash, error) {
	repository, err := git.PlainOpen(gitops.Path)

	if err != nil {
		return plumbing.Hash{}, err
	}

	remotes, err := repository.Remotes()

	if err != nil {
		return plumbing.Hash{}, err
	}

	remote := git.NewRemote(repository.Storer, remotes[0].Config())

	if err != nil {
		return plumbing.Hash{}, err
	}

	refs, err := remote.List(&git.ListOptions{
		Auth: gitops.AuthResolved,
	})

	if len(refs) > 0 {
		return refs[0].Hash(), nil
	} else {
		return plumbing.Hash{}, errors.New("refs empty list")
	}
}

func (gitops *Gitops) PathExists() bool {
	_, err := os.Stat(fmt.Sprintf("%s/%s", gitops.Path, gitops.DirectoryPath))
	return os.IsExist(err)
}

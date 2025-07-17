package internal

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"os"
	"path"
	"time"
)

var (
	ErrPollingInterval = errors.New("sleep interval didn't pass")
)

type Git struct {
	Repository string
	Revision   string
	Directory  string
	LogPath    string
	Auth       *Auth `json:"-"`
}

func NewGit(definition *v1.GitopsDefinition, logpath string) *Git {
	directory := fmt.Sprintf("/tmp/%s", path.Base(definition.Spec.RepoURL))

	return &Git{
		Repository: definition.Spec.RepoURL,
		Revision:   definition.Spec.Revision,
		Directory:  directory,
		LogPath:    logpath,
		Auth:       NewAuth(),
	}
}

func (g *Git) Fetch() (*object.Commit, error) {
	if _, err := git.PlainOpen(g.Directory); err != nil {
		err = g.Clone()

		if err != nil {
			if errors.Is(err, git.ErrRepositoryAlreadyExists) {
				return g.Pull()
			}

			return nil, err
		}

		return g.Pull()
	} else {
		return g.Pull()
	}
}

func (g *Git) Clone() error {
	file, err := g.LogOpen()

	if err != nil {
		return err
	}

	defer g.LogClose(file)

	_, err = git.PlainClone(g.Directory, false, &git.CloneOptions{
		URL:           g.Repository,
		Progress:      file,
		ReferenceName: plumbing.NewBranchReferenceName(g.Revision),
		Auth:          g.Auth.Auth,
	})

	if err != nil {
		return err
	}

	return nil
}

func (g *Git) Commit(message string) error {
	file, err := g.LogOpen()
	if err != nil {
		return err
	}
	defer g.LogClose(file)

	repository, err := git.PlainOpen(g.Directory)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	workTree, err := repository.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get working tree: %w", err)
	}

	err = workTree.AddWithOptions(&git.AddOptions{
		All: true,
	})
	if err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	status, err := workTree.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if status.IsClean() {
		return fmt.Errorf("no changes to commit")
	}

	commit, err := workTree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "simplecontainer-bot",
			Email: "bot@simplecontainer.io",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	fmt.Printf("Commit created: %s\n", commit.String())

	return nil
}

func (g *Git) CommitFiles(message string, files []string) error {
	file, err := g.LogOpen()
	if err != nil {
		return err
	}
	defer g.LogClose(file)

	repository, err := git.PlainOpen(g.Directory)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	workTree, err := repository.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get working tree: %w", err)
	}

	for _, file := range files {
		_, err = workTree.Add(file)
		if err != nil {
			return fmt.Errorf("failed to add file %s: %w", file, err)
		}
	}

	status, err := workTree.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if status.IsClean() {
		return fmt.Errorf("no changes to commit")
	}

	commit, err := workTree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "simplecontainer-bot",
			Email: "bot@simplecontainer.io",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	fmt.Printf("Commit created: %s\n", commit.String())

	return nil
}

func (g *Git) Push() error {
	file, err := g.LogOpen()
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer g.LogClose(file)

	repository, err := git.PlainOpen(g.Directory)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	err = repository.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       g.Auth.Auth,
		Force:      false,
		Progress:   file,
	})
	if err != nil {
		// Handle specific Git errors
		if err == git.NoErrAlreadyUpToDate {
			fmt.Println("Repository is already up to date")
			return nil
		}
		return fmt.Errorf("failed to push to remote: %w", err)
	}

	fmt.Println("Successfully pushed to origin")
	return nil
}

// Alternative: Push to specific remote and branch
func (g *Git) PushToRemote(remoteName, refSpec string) error {
	file, err := g.LogOpen()
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer g.LogClose(file)

	repository, err := git.PlainOpen(g.Directory)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	pushOptions := &git.PushOptions{
		RemoteName: remoteName,
		Auth:       g.Auth.Auth,
		Force:      false,
		Progress:   file,
	}

	if refSpec != "" {
		pushOptions.RefSpecs = []config.RefSpec{config.RefSpec(refSpec)}
	}

	err = repository.Push(pushOptions)
	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			fmt.Printf("Repository is already up to date with %s\n", remoteName)
			return nil
		}
		return fmt.Errorf("failed to push to %s: %w", remoteName, err)
	}

	fmt.Printf("Successfully pushed to %s\n", remoteName)
	return nil
}

func (g *Git) Pull() (*object.Commit, error) {
	file, err := g.LogOpen()

	if err != nil {
		return nil, err
	}

	defer g.LogClose(file)

	repository, err := git.PlainOpen(g.Directory)

	if err != nil {
		return nil, err
	}

	worktree, _ := repository.Worktree()

	var ref *plumbing.Reference

	err = worktree.Pull(&git.PullOptions{
		RemoteName:    "origin",
		Auth:          g.Auth.Auth,
		SingleBranch:  true,
		Force:         true,
		Progress:      file,
		ReferenceName: plumbing.NewBranchReferenceName(g.Revision),
	})

	ref, err = repository.Head()

	if err != nil {
		return nil, err
	}

	return repository.CommitObject(ref.Hash())
}

func (g *Git) RemoteHead() (plumbing.Hash, error) {
	repository, err := git.PlainOpen(g.Directory)

	if err != nil {
		return plumbing.Hash{}, nil
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
		Auth: g.Auth.Auth,
	})

	if len(refs) > 0 {
		return refs[0].Hash(), nil
	} else {
		return plumbing.Hash{}, errors.New("refs empty list")
	}
}

func (g *Git) LogOpen() (*os.File, error) {
	file, err := os.OpenFile(g.LogPath, os.O_RDWR, 0644)

	if err != nil {
		return nil, err
	}

	return file, nil
}

func (g *Git) LogClose(file *os.File) error {
	return file.Close()
}

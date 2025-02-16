package internal

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"os"
	"path"
)

var (
	ErrPoolingInterval = errors.New("sleep interval didn't pass")
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

package reconcile

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/simplecontainer/smr/implementations/gitops/gitops"
	"github.com/simplecontainer/smr/implementations/gitops/shared"
	"github.com/simplecontainer/smr/pkg/definitions"
	"os"
	"path/filepath"
)

func Clone(gitopsObj *gitops.Gitops, auth transport.AuthMethod, path string) (plumbing.Hash, error) {
	var repository *git.Repository
	var err error

	if _, err = os.Stat(path); errors.Is(err, os.ErrNotExist) {
		repository, err = git.PlainClone(path, false, &git.CloneOptions{
			URL:           gitopsObj.RepoURL,
			Progress:      os.Stdout,
			ReferenceName: plumbing.NewBranchReferenceName(gitopsObj.Revision),
			Auth:          auth,
		})

		if err != nil {
			return plumbing.Hash{}, err
		}
	} else {
		repository, err = git.PlainOpen(path)

		if err != nil {
			return plumbing.Hash{}, err
		}
	}

	worktree, _ := repository.Worktree()

	err = worktree.Pull(&git.PullOptions{
		RemoteName:    "origin",
		Auth:          auth,
		SingleBranch:  true,
		Force:         true,
		ReferenceName: plumbing.NewBranchReferenceName(gitopsObj.Revision),
	})

	if err != nil {
		if err.Error() != "already up-to-date" {
			return plumbing.Hash{}, err
		}
	}

	var ref *plumbing.Reference
	var commit *object.Commit

	ref, _ = repository.Head()
	commit, err = repository.CommitObject(ref.Hash())

	if commit == nil {
		return plumbing.Hash{}, err
	}

	return commit.ID(), nil
}

func SortFiles(gitopsObj *gitops.Gitops, localPath string, shared *shared.Shared) ([]map[string]string, error) {
	entries, err := os.ReadDir(fmt.Sprintf("%s%s", localPath, gitopsObj.DirectoryPath))

	if err != nil {
		return nil, err
	}

	orderedByDependencies := make([]map[string]string, 0)

	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".yaml" {
			definition := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", localPath, gitopsObj.DirectoryPath, e.Name()))
			if err != nil {
				return nil, err
			}

			data := make(map[string]interface{})

			err = json.Unmarshal([]byte(definition), &data)
			if err != nil {
				return nil, err
			}

			position := -1

			for index, orderedEntry := range orderedByDependencies {
				deps := shared.Manager.RelationRegistry.GetDependencies(orderedEntry["kind"])

				for _, dp := range deps {
					if data["kind"].(string) == dp {
						position = index
					}
				}
			}

			if data["kind"] != nil {
				if position != -1 {
					orderedByDependencies = append(orderedByDependencies[:position+1], orderedByDependencies[position:]...)
					orderedByDependencies[position] = map[string]string{"name": e.Name(), "kind": data["kind"].(string)}
				} else {
					orderedByDependencies = append(orderedByDependencies, map[string]string{"name": e.Name(), "kind": data["kind"].(string)})
				}
			}
		}
	}

	return orderedByDependencies, nil
}

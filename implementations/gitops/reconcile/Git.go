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

func Clone(gitopsObj *gitops.Gitops, auth transport.AuthMethod, localPath string) (plumbing.Hash, error) {
	if _, err := os.Stat(localPath); errors.Is(err, os.ErrNotExist) {
		_, err = git.PlainClone(localPath, false, &git.CloneOptions{
			URL:      gitopsObj.RepoURL,
			Progress: os.Stdout,
			Auth:     auth,
		})

		fmt.Println(err)
		fmt.Println(auth.String())
		fmt.Println(auth.Name())

		if err != nil {
			return plumbing.Hash{}, err
		}
	}

	var r *git.Repository
	var err error

	var commit *object.Commit

	r, err = git.PlainOpen(localPath)

	if err != nil {
		return plumbing.Hash{}, err
	}

	w, _ := r.Worktree()

	_ = w.Pull(&git.PullOptions{RemoteName: "origin", Auth: auth})

	ref, _ := r.Head()
	commit, err = r.CommitObject(ref.Hash())

	if commit == nil {
		return plumbing.Hash{}, err
	}

	return commit.Hash, err
}

func SortFiles(gitopsObj *gitops.Gitops, localPath string, shared *shared.Shared) ([]map[string]string, error) {
	entries, err := os.ReadDir(fmt.Sprintf("%s%s", localPath, gitopsObj.DirectoryPath))

	if err != nil {
		return nil, err
	}

	orderedByDependencies := make([]map[string]string, 0)

	for _, e := range entries {
		if filepath.Ext(e.Name()) == "yaml" {
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

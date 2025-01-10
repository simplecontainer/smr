package implementation

import (
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/pkg/definitions"
	"github.com/simplecontainer/smr/pkg/relations"
	"os"
	"path/filepath"
)

func (gitops *Gitops) Definitions(relations *relations.RelationRegistry) ([]FileKind, error) {
	entries, err := os.ReadDir(filepath.Clean(fmt.Sprintf("%s/%s", gitops.Path, gitops.DirectoryPath)))

	if err != nil {
		return nil, err
	}

	orderedByDependencies := make([]FileKind, 0)

	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".yaml" {
			var definition []byte
			definition, err = definitions.ReadFile(fmt.Sprintf("%s/%s/%s", gitops.Path, gitops.DirectoryPath, e.Name()))

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
				deps := relations.GetDependencies(orderedEntry.Kind)

				for _, dp := range deps {
					if data["kind"].(string) == dp {
						position = index
					}
				}
			}

			if data["kind"] != nil {
				if position != -1 {
					orderedByDependencies = append(orderedByDependencies[:position+1], orderedByDependencies[position:]...)
					orderedByDependencies[position] = FileKind{
						File: e.Name(),
						Kind: data["kind"].(string),
					}
				} else {
					orderedByDependencies = append(orderedByDependencies, FileKind{
						File: e.Name(),
						Kind: data["kind"].(string),
					})
				}
			}
		}
	}

	return orderedByDependencies, nil
}

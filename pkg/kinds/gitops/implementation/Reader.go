package implementation

import (
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/pkg/definitions"
	"github.com/simplecontainer/smr/pkg/relations"
	"os"
	"path/filepath"
)

func (gitops *Gitops) Definitions(relations *relations.RelationRegistry) ([]map[string]string, error) {
	entries, err := os.ReadDir(filepath.Clean(fmt.Sprintf("%s/%s", gitops.Path, gitops.DirectoryPath)))

	if err != nil {
		return nil, err
	}

	orderedByDependencies := make([]map[string]string, 0)

	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".yaml" {
			definition := definitions.ReadFile(fmt.Sprintf("%s/%s/%s", gitops.Path, gitops.DirectoryPath, e.Name()))
			data := make(map[string]interface{})

			err = json.Unmarshal([]byte(definition), &data)

			if err != nil {
				return nil, err
			}

			position := -1

			for index, orderedEntry := range orderedByDependencies {
				deps := relations.GetDependencies(orderedEntry["kind"])

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

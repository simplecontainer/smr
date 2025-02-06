package implementation

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/packer"
	"github.com/simplecontainer/smr/pkg/relations"
	"os"
	"path/filepath"
)

func (gitops *Gitops) ReadDefinitions(relations *relations.RelationRegistry) ([]*common.Request, error) {
	entries, err := os.ReadDir(filepath.Clean(fmt.Sprintf("%s/%s", gitops.Path, gitops.DirectoryPath)))

	if err != nil {
		return nil, err
	}

	var requests []*common.Request
	orderedByDependencies := make([]*common.Request, 0)

	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".yaml" {
			var definition []byte
			definition, err = packer.ReadYAMLFile(fmt.Sprintf("%s/%s/%s", gitops.Path, gitops.DirectoryPath, e.Name()))

			if err != nil {
				return nil, err
			}

			requests, err = packer.Parse(definition)

			if err != nil {
				return nil, err
			}

			for _, request := range requests {
				position := -1

				for index, orderedEntry := range orderedByDependencies {
					deps := relations.GetDependencies(orderedEntry.Kind)

					for _, dp := range deps {
						if request.Definition.GetKind() == dp {
							position = index
						}
					}
				}

				if request.Definition.GetKind() != "" {
					if position != -1 {
						orderedByDependencies = append(orderedByDependencies[:position+1], orderedByDependencies[position:]...)
						orderedByDependencies[position] = request
					} else {
						orderedByDependencies = append(orderedByDependencies, request)
					}
				}
			}
		}
	}

	return orderedByDependencies, nil
}

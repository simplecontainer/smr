package packer

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/relations"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func Read(path string, relations *relations.RelationRegistry) (*Pack, error) {
	pack := New()

	entries, err := os.ReadDir(fmt.Sprintf("%s/definitions", filepath.Clean(path)))

	if err != nil {
		return pack, err
	}

	var packID []byte
	packID, err = ReadYAMLFile(fmt.Sprintf("%s/Pack.yaml", filepath.Clean(path)))

	if err != nil {
		return pack, err
	}

	err = yaml.Unmarshal(packID, pack)

	if err != nil {
		return pack, err
	}

	var requests []*common.Request
	ordered := make([]*common.Request, 0)

	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".yaml" {
			var definition []byte
			definition, err = ReadYAMLFile(fmt.Sprintf("%s/definitions/%s", path, e.Name()))

			if err != nil {
				return pack, err
			}

			requests, err = Parse(definition)

			if err != nil {
				return pack, err
			}

			for _, request := range requests {
				position := -1

				for index, element := range ordered {
					deps := relations.GetDependencies(element.Kind)

					for _, dp := range deps {
						if request.Definition.GetKind() == dp {
							position = index
						}
					}
				}

				if request.Definition.GetKind() != "" {
					if position != -1 {
						ordered = append(ordered[:position+1], ordered[position:]...)
						ordered[position] = request
					} else {
						ordered = append(ordered, request)
					}
				}
			}
		}
	}

	pack.Definitions = ordered

	return pack, nil
}

package docker

import "github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"

func (container *Docker) HasDependencyOn(kind string, group string, identifier string, runtime *types.Runtime) bool {
	for _, format := range runtime.ObjectDependencies {
		if format.Identifier == identifier && format.Group == group && format.Kind == kind {
			return true
		}
	}

	return false
}

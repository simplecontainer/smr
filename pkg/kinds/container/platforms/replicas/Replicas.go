package replicas

import (
	"github.com/simplecontainer/smr/pkg/configuration"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/container/registry"
	"github.com/simplecontainer/smr/pkg/kinds/container/status"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/static"
)

func New(nodeID uint64, nodes []*node.Node) *Replicas {
	cluster := make([]uint64, 0)

	for _, n := range nodes {
		cluster = append(cluster, n.NodeID)
	}

	return &Replicas{
		NodeID:  nodeID,
		Create:  []uint64{0},
		Destroy: []uint64{0},
		Cluster: cluster,
	}
}

func (replicas *Replicas) GenerateContainers(registry *registry.Registry, definition *v1.ContainerDefinition, config *configuration.Configuration) ([]platforms.IContainer, []platforms.IContainer, error) {
	create, destroy := replicas.GetContainersIndexes(registry, definition)

	createContainers := make([]platforms.IContainer, 0)
	destroyContainers := make([]platforms.IContainer, 0)

	for _, index := range create {
		generatedName := registry.NameReplica(definition.Meta.Group, definition.Meta.Name, index)

		newContainer, err := platforms.New(static.PLATFORM_DOCKER, generatedName, config, registry.ChangeC, definition)

		if err != nil {
			return createContainers, destroyContainers, err
		}

		existing := registry.FindLocal(newContainer.GetGroup(), newContainer.GetGeneratedName())

		if existing == nil || existing.GetStatus().GetCategory() == status.CATEGORY_END {
			// Since container already exists in local registry don't recreate it - it's good
			createContainers = append(createContainers, newContainer)
		}
	}

	for _, index := range destroy {
		generatedName := registry.NameReplica(definition.Meta.Group, definition.Meta.Name, index)
		existing := registry.FindLocal(definition.Meta.Group, generatedName)

		if existing != nil {
			destroyContainers = append(destroyContainers, existing)
		}
	}

	return createContainers, destroyContainers, nil
}

func (replicas *Replicas) RemoveContainers(registry *registry.Registry, definition *v1.ContainerDefinition) ([]platforms.IContainer, error) {
	destroy, _ := replicas.GetContainersIndexes(registry, definition)

	destroyContainers := make([]platforms.IContainer, 0)

	for _, index := range destroy {
		generatedName := registry.NameReplica(definition.Meta.Group, definition.Meta.Name, index)
		existing := registry.FindLocal(definition.Meta.Group, generatedName)

		if existing != nil {
			destroyContainers = append(destroyContainers, existing)
		}
	}

	return destroyContainers, nil
}

func (replicas *Replicas) GetContainersIndexes(registry *registry.Registry, definition *v1.ContainerDefinition) ([]uint64, []uint64) {
	replicas.Recalculate(definition.Spec.Container.Spread, definition.Spec.Container.Replicas, registry.GetIndexes(definition.Meta.Group, definition.Meta.Name))
	return replicas.Create, replicas.Destroy
}

func (replicas *Replicas) Recalculate(spread v1.ContainerSpread, replicasDefined uint64, existingIndexes []uint64) {
	replicas.Create, replicas.Destroy = replicas.GetReplicaNumbers(spread, replicasDefined, existingIndexes)
}

func (replicas *Replicas) GetReplicaNumbers(spread v1.ContainerSpread, replicasDefined uint64, existingIndexes []uint64) ([]uint64, []uint64) {
	switch spread.Spread {
	case platforms.SPREAD_SPECIFIC:
		return Specific(replicasDefined, existingIndexes, spread.Agents, replicas.NodeID)
	case platforms.SPREAD_UNIFORM:
		return Uniform(replicasDefined, existingIndexes, replicas.Cluster, replicas.NodeID)
	default:
		return Default(replicasDefined, existingIndexes)
	}
}

func Default(replicasNumber uint64, existingIndexes []uint64) ([]uint64, []uint64) {
	var create = make([]uint64, 0)
	var destroy = make([]uint64, 0)

	for replicas := uint64(1); replicas <= replicasNumber; replicas++ {
		create = append(create, replicas)
	}

	if len(create) < len(existingIndexes) {
		destroy = existingIndexes[len(create):]
	}

	return create, destroy
}
func Uniform(replicasWanted uint64, existingIndexes []uint64, cluster []uint64, member uint64) ([]uint64, []uint64) {
	var create = make([][]uint64, 0)
	var destroy = make([]uint64, 0)

	replicas := make([]uint64, replicasWanted)
	for i := uint64(0); i < replicasWanted; i++ {
		replicas[i] = i + uint64(1)
	}

	create = ChunkSlice(replicas, len(cluster))

	if len(create[member-1]) < len(existingIndexes) {
		for i, existing := range existingIndexes {
			preserve := false

			for _, creating := range create[member-1] {
				if creating == existing {
					preserve = true
				}
			}

			if !preserve {
				destroy = append(destroy, existingIndexes[i])
			}
		}
	}

	return create[member-1], destroy
}
func Specific(replicasWanted uint64, existingIndexes []uint64, nodes []uint64, member uint64) ([]uint64, []uint64) {
	var create = make([][]uint64, 0)
	var destroy = make([]uint64, 0)

	replicas := make([]uint64, replicasWanted)
	for i := uint64(0); i < replicasWanted; i++ {
		replicas[i] = i + uint64(1)
	}

	create = ChunkSlice(replicas, len(nodes))

	if len(create[member-1]) < len(existingIndexes) {
		for i, existing := range existingIndexes {
			preserve := false

			for _, creating := range create[member-1] {
				if creating == existing {
					preserve = true
				}
			}

			if !preserve {
				destroy = append(destroy, existingIndexes[i])
			}
		}
	}

	return create[member-1], destroy
}

func ChunkSlice(slice []uint64, numChunks int) [][]uint64 {
	chunkSize := len(slice) / numChunks
	extra := len(slice) % numChunks

	var result [][]uint64
	start := 0

	for i := 0; i < numChunks; i++ {
		end := start + chunkSize
		if i < extra {
			end++
		}

		result = append(result, slice[start:end])
		start = end
	}

	return result
}

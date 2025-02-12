package replicas

import (
	"github.com/simplecontainer/smr/pkg/configuration"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/static"
	"slices"
	"sort"
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

func (replicas *Replicas) GenerateContainers(registry platforms.Registry, definition *v1.ContainersDefinition, config *configuration.Configuration) ([]platforms.IContainer, []platforms.IContainer, []platforms.IContainer, error) {
	create, destroy := replicas.GetContainersIndexes(registry, definition)

	createContainers := make([]platforms.IContainer, 0)
	updateContainers := make([]platforms.IContainer, 0)
	destroyContainers := make([]platforms.IContainer, 0)

	for _, index := range create {
		generatedName := registry.NameReplica(definition.Meta.Group, definition.Meta.Name, index)

		newContainer, err := platforms.New(static.PLATFORM_DOCKER, generatedName, config, definition)

		if err != nil {
			return createContainers, updateContainers, destroyContainers, err
		}

		existing := registry.FindLocal(newContainer.GetGroup(), newContainer.GetGeneratedName())

		if existing == nil {
			createContainers = append(createContainers, newContainer)
		} else {
			updateContainers = append(updateContainers, newContainer)
		}
	}

	for _, index := range destroy {
		generatedName := registry.NameReplica(definition.Meta.Group, definition.Meta.Name, index)
		existing := registry.FindLocal(definition.Meta.Group, generatedName)

		if existing != nil {
			destroyContainers = append(destroyContainers, existing)
		}
	}

	return createContainers, updateContainers, destroyContainers, nil
}

func (replicas *Replicas) RemoveContainers(registry platforms.Registry, definition *v1.ContainersDefinition) ([]platforms.IContainer, error) {
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

func (replicas *Replicas) GetContainersIndexes(registry platforms.Registry, definition *v1.ContainersDefinition) ([]uint64, []uint64) {
	if definition.Spec.Spread.Spread == "" {
		// No spread so create only for node who sourced the object
		replicas.Recalculate(v1.ContainersSpread{
			Spread: "specific",
			Agents: []uint64{definition.GetRuntime().GetNode()},
		}, definition.Spec.Replicas, registry.GetIndexes(definition.Meta.Group, definition.Meta.Name))
	} else {
		replicas.Recalculate(definition.Spec.Spread, definition.Spec.Replicas, registry.GetIndexes(definition.Meta.Group, definition.Meta.Name))
	}

	return replicas.Create, replicas.Destroy
}

func (replicas *Replicas) Recalculate(spread v1.ContainersSpread, replicasDefined uint64, existingIndexes []uint64) {
	replicas.Create, replicas.Destroy = replicas.GetReplicaNumbers(spread, replicasDefined, existingIndexes)
}

func (replicas *Replicas) GetReplicaNumbers(spread v1.ContainersSpread, replicasDefined uint64, existingIndexes []uint64) ([]uint64, []uint64) {
	switch spread.Spread {
	case platforms.SPREAD_SPECIFIC:
		return Specific(replicasDefined, existingIndexes, spread.Agents, replicas.NodeID)
	case platforms.SPREAD_UNIFORM:
		return Uniform(replicasDefined, existingIndexes, replicas.Cluster, replicas.NodeID)
	default:
		return Specific(replicasDefined, existingIndexes, spread.Agents, replicas.NodeID)
	}
}

func Uniform(replicasWanted uint64, existingIndexes []uint64, cluster []uint64, member uint64) ([]uint64, []uint64) {
	var create = make([][]uint64, 0)
	var destroy = make([]uint64, 0)

	replicas := make([]uint64, replicasWanted)
	for i := uint64(0); i < replicasWanted; i++ {
		replicas[i] = i + uint64(1)
	}

	create = ChunkSlice(replicas, len(cluster))

	if len(create[member-1]) <= len(existingIndexes) {
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

	if slices.Contains(nodes, member) {
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i] < nodes[j]
		})

		normalizedNodes := make(map[uint64]uint64)
		normalizedMember := uint64(0)

		for i, node := range nodes {
			x := node - uint64(i)
			normalizedNodes[uint64(i)] = nodes[i] - x + 1

			if node == member {
				normalizedMember = uint64(i)
			}
		}

		replicas := make([]uint64, replicasWanted)
		for i := uint64(0); i < replicasWanted; i++ {
			replicas[i] = i + uint64(1)
		}

		create = ChunkSlice(replicas, len(nodes))

		if len(create[normalizedMember]) <= len(existingIndexes) {
			for i, existing := range existingIndexes {
				preserve := false

				for _, creating := range create[normalizedMember] {
					if creating == existing {
						preserve = true
					}
				}

				if !preserve {
					destroy = append(destroy, existingIndexes[i])
				}
			}
		}

		return create[normalizedMember], destroy
	}

	return []uint64{}, []uint64{}
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

package replicas

import (
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/static"
	"slices"
	"sort"
	"strings"
)

func NewReplica(shared *shared.Shared, agent string, definition *v1.ContainerDefinition, changelog diff.Changelog) *Replicas {
	return &Replicas{
		Definition:      definition,
		Shared:          shared,
		NodeID:          shared.Manager.Config.KVStore.Node,
		Distributed:     NewDistributed(shared.Manager.Config.KVStore.Node, definition.Meta.Group, definition.Meta.Name),
		ChangeLog:       changelog,
		ExistingIndexes: shared.Registry.GetIndexes(definition.Meta.Group, definition.Meta.Name),
		Agent:           agent,
	}
}

func (replicas *Replicas) HandleReplica(user *authentication.User, clstr []string) (*Distributed, error) {
	dr := NewDistributed(replicas.NodeID, replicas.Definition.Meta.Group, replicas.Definition.Meta.Name)
	err := dr.Load(replicas.Shared.Client.Get(user.Username), user)

	dr.Replicas[replicas.NodeID].Numbers.Create, dr.Replicas[replicas.NodeID].Numbers.Destroy, dr.Replicas[replicas.NodeID].Numbers.Existing = replicas.GetReplicaNumbers(dr, replicas.Definition.Spec.Container.Spread, replicas.Definition.Spec.Container.Replicas, replicas.ExistingIndexes, uint64(len(clstr)))
	dr.Clear(replicas.NodeID)

	create := dr.Replicas[replicas.NodeID].Numbers.Create
	destroy := dr.Replicas[replicas.NodeID].Numbers.Destroy

	// Destroy from the end to start
	if len(destroy) > 0 {
		for _, v := range destroy {
			name, _ := replicas.Shared.Registry.NameReplicas(replicas.Definition.Meta.Group, replicas.Definition.Meta.Name, v)
			existingContainer := replicas.Shared.Registry.FindLocal(replicas.Definition.Meta.Group, name)

			if existingContainer != nil {
				dr.Replicas[replicas.NodeID].Delete(existingContainer.GetGroup(), existingContainer.GetGeneratedName())
			}
		}
	}

	// Create from the start to the end
	for _, v := range create {
		name, _ := replicas.Shared.Registry.NameReplicas(replicas.Definition.Meta.Group, replicas.Definition.Meta.Name, v)
		existingContainer := replicas.Shared.Registry.FindLocal(replicas.Definition.Meta.Group, name)

		if existingContainer != nil {
			var onlyReplicaChange = false

			if len(replicas.ChangeLog) == 1 {
				for _, change := range replicas.ChangeLog {
					if strings.Join(change.Path, ":") == "Spec:Container:Replicas" {
						if change.Type == "update" {
							onlyReplicaChange = true
						}
					}
				}
			}

			if onlyReplicaChange {
				continue
			}
		}

		var containerObj platforms.IContainer

		containerObj, err = platforms.New(static.PLATFORM_DOCKER, name, replicas.Shared.Manager.Config, replicas.Shared.Registry.ChangeC, replicas.Definition)

		if err != nil {
			return nil, err
		}

		if existingContainer == nil {
			replicas.Shared.Registry.AddOrUpdate(replicas.Definition.Meta.Group, name, containerObj)
		} else {
			if len(replicas.ChangeLog) > 0 {
				replicas.Shared.Registry.AddOrUpdate(replicas.Definition.Meta.Group, name, containerObj)
			}
		}

		dr.Replicas[replicas.NodeID].Add(replicas.Definition.Meta.Group, name)
	}

	err = dr.Save(replicas.Shared.Client.Get(user.Username), user)

	if err != nil {
		return nil, err
	}

	return dr, nil
}

func (replicas *Replicas) GetReplica(user *authentication.User, clstr []string) (*Distributed, error) {
	dr := NewDistributed(replicas.NodeID, replicas.Definition.Meta.Group, replicas.Definition.Meta.Name)
	dr.Load(replicas.Shared.Client.Get(user.Username), user)

	_, _, dr.Replicas[replicas.NodeID].Numbers.Existing = replicas.GetReplicaNumbers(dr, replicas.Definition.Spec.Container.Spread, replicas.Definition.Spec.Container.Replicas, replicas.ExistingIndexes, uint64(len(clstr)))

	for _, v := range dr.Replicas[replicas.NodeID].Numbers.Existing {
		name, _ := replicas.Shared.Registry.NameReplicas(replicas.Definition.Meta.Group, replicas.Definition.Meta.Name, v)
		existingContainer := replicas.Shared.Registry.FindLocal(replicas.Definition.Meta.Group, name)

		if existingContainer != nil {
			dr.Replicas[replicas.NodeID].AddExisting(replicas.Definition.Meta.Group, name)
		}
	}

	return dr, nil
}

func (replicas *Replicas) GetReplicaNumbers(dr *Distributed, spread v1.ContainerSpread, replicasNumber uint64, existingIndexes []uint64, clusterSize uint64) ([]uint64, []uint64, []uint64) {
	switch replicas.Spread.Spread {
	case platforms.SPREAD_SPECIFIC:
		if dr.Replicas[replicas.NodeID] == nil {
			dr.Replicas[replicas.NodeID] = NewScoped()
		}

		dr.Replicas[replicas.NodeID].Numbers.Create, dr.Replicas[replicas.NodeID].Numbers.Destroy, dr.Replicas[replicas.NodeID].Numbers.Existing = Specific(replicasNumber, existingIndexes, spread.Agents, replicas.NodeID)
		break
	case platforms.SPREAD_UNIFORM:
		if dr.Replicas[replicas.NodeID] == nil {
			dr.Replicas[replicas.NodeID] = NewScoped()
		}

		dr.Replicas[replicas.NodeID].Numbers.Create, dr.Replicas[replicas.NodeID].Numbers.Destroy, dr.Replicas[replicas.NodeID].Numbers.Existing = Uniform(replicasNumber, existingIndexes, clusterSize, replicas.NodeID)
		break
	default:
		if replicas.Agent != replicas.Shared.Manager.Config.Node {
			dr.Replicas[replicas.NodeID].Numbers.Create = []uint64{}
			dr.Replicas[replicas.NodeID].Numbers.Destroy = []uint64{}
			dr.Replicas[replicas.NodeID].Numbers.Existing = []uint64{}
		} else {
			dr.Replicas[replicas.NodeID].Numbers.Create, dr.Replicas[replicas.NodeID].Numbers.Destroy, dr.Replicas[replicas.NodeID].Numbers.Existing = Default(replicasNumber, existingIndexes)
		}
		break
	}

	return dr.Replicas[replicas.NodeID].Numbers.Create, dr.Replicas[replicas.NodeID].Numbers.Destroy, dr.Replicas[replicas.NodeID].Numbers.Existing
}

func Default(replicasNumber uint64, existingIndexes []uint64) ([]uint64, []uint64, []uint64) {
	var create = make([]uint64, 0)
	var destroy = make([]uint64, 0)

	for replicas := uint64(1); replicas <= replicasNumber; replicas++ {
		create = append(create, replicas)
	}

	return create, destroy, existingIndexes
}

func Uniform(replicasNumber uint64, existingIndexes []uint64, clusterSize uint64, member uint64) ([]uint64, []uint64, []uint64) {
	var create = make([]uint64, 0)
	var destroy = make([]uint64, 0)

	replicasScoped := replicasNumber / clusterSize

	for replicas := 1 * member; replicas <= replicasScoped*member; replicas++ {
		create = append(create, replicas)
	}

	if clusterSize == member && replicasNumber%clusterSize != 0 {
		create = append(create, replicasNumber)
	}

	if len(create) < len(existingIndexes) {
		destroy = existingIndexes[len(create):]
	}

	return create, destroy, existingIndexes
}

func Specific(replicasNumber uint64, existingIndexes []uint64, nodes []uint64, member uint64) ([]uint64, []uint64, []uint64) {
	var create = make([]uint64, 0)
	var destroy = make([]uint64, 0)

	if slices.Contains(nodes, member) {
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i] < nodes[j]
		})

		nodeCount := uint64(len(nodes))
		replicasScoped := replicasNumber / nodeCount
		normalizedNodes := make(map[uint64]uint64)

		// Normalize node ID -> eg. 2,5,7 -> 1,2,3
		for i, node := range nodes {
			x := node - uint64(i)
			normalizedNodes[member] = nodes[i] - x
		}

		for replicas := 1 * normalizedNodes[member]; replicas <= replicasScoped*normalizedNodes[member]; replicas++ {
			create = append(create, replicas)
		}

		if nodeCount == normalizedNodes[member] && replicasNumber%nodeCount != 0 {
			create = append(create, replicasNumber)
		}

		if len(create) < len(existingIndexes) {
			destroy = existingIndexes[len(create):]
		}
	}

	return create, destroy, existingIndexes
}

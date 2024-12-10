package replicas

import (
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
	"strings"
)

func (replicas *Replicas) HandleReplica(shared *shared.Shared, user *authentication.User, clstr []string, containerDefinition *v1.ContainerDefinition, changelog diff.Changelog) (*DistributedReplicas, error) {
	dr := NewDistributed(replicas.NodeID, containerDefinition.Meta.Group, containerDefinition.Meta.Name)
	err := dr.Load(shared.Client.Get(user.Username), user)

	if err != nil {
		logger.Log.Error(err.Error())
	}

	dr.Replicas[replicas.NodeID].Numbers.Create, dr.Replicas[replicas.NodeID].Numbers.Destroy, dr.Replicas[replicas.NodeID].Numbers.Existing = replicas.GetReplicaNumbers(dr, containerDefinition.Spec.Container.Spread, replicas.Replicas, replicas.ExistingIndexes, uint64(len(clstr)))
	dr.Clear(replicas.NodeID)

	create := dr.Replicas[replicas.NodeID].Numbers.Create
	destroy := dr.Replicas[replicas.NodeID].Numbers.Destroy

	// Destroy from the end to start
	if len(destroy) > 0 {
		for _, v := range destroy {
			name, _ := shared.Registry.NameReplicas(containerDefinition.Meta.Group, containerDefinition.Meta.Name, v)
			existingContainer := shared.Registry.FindLocal(containerDefinition.Meta.Group, name)

			if existingContainer != nil {
				dr.Replicas[replicas.NodeID].Delete(existingContainer.GetGroup(), existingContainer.GetGeneratedName())
			}
		}
	}

	// Create from the start to the end
	for _, v := range create {
		name, _ := shared.Registry.NameReplicas(containerDefinition.Meta.Group, containerDefinition.Meta.Name, v)
		existingContainer := shared.Registry.FindLocal(containerDefinition.Meta.Group, name)

		if existingContainer != nil {
			var onlyReplicaChange = false

			if len(changelog) == 1 {
				for _, change := range changelog {
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

		containerObj, err = platforms.New(static.PLATFORM_DOCKER, name, shared.Manager.Config, shared.Registry.ChangeC, containerDefinition)

		if err != nil {
			return nil, err
		}

		if existingContainer == nil {
			shared.Registry.AddOrUpdate(replicas.Group, name, containerObj)
		} else {
			if replicas.Changed {
				shared.Registry.AddOrUpdate(replicas.Group, name, containerObj)
			}
		}

		dr.Replicas[replicas.NodeID].Add(replicas.Group, name)
	}

	err = dr.Save(shared.Client.Get(user.Username), user)

	if err != nil {
		return nil, err
	}

	return dr, nil
}

func (replicas *Replicas) GetReplica(shared *shared.Shared, user *authentication.User, containerDefinition v1.ContainerDefinition, clstr []string) ([]R, error) {
	containers := make([]R, 0)

	dr := NewDistributed(replicas.NodeID, containerDefinition.Meta.Group, containerDefinition.Meta.Name)
	dr.Load(shared.Client.Get(user.Username), user)

	_, _, dr.Replicas[replicas.NodeID].Numbers.Existing = replicas.GetReplicaNumbers(dr, containerDefinition.Spec.Container.Spread, replicas.Replicas, replicas.ExistingIndexes, uint64(len(clstr)))

	for i := range dr.Replicas[replicas.NodeID].Numbers.Existing {
		name, _ := shared.Registry.NameReplicas(containerDefinition.Meta.Group, containerDefinition.Meta.Name, uint64(i))
		containerObj := shared.Registry.FindLocal(containerDefinition.Meta.Group, name)

		if containerObj != nil {
			containers = append(containers, R{replicas.Group, name})
		}
	}

	return containers, nil
}

func (replicas *Replicas) GetReplicaNumbers(dr *DistributedReplicas, spread v1.ContainerSpread, replicasNumber uint64, existingIndexes []uint64, clusterSize uint64) ([]uint64, []uint64, []uint64) {
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
		dr.Replicas[replicas.NodeID].Numbers.Create, dr.Replicas[replicas.NodeID].Numbers.Destroy, dr.Replicas[replicas.NodeID].Numbers.Existing = Default(replicasNumber, existingIndexes, 1, 1)
		break
	}

	return dr.Replicas[replicas.NodeID].Numbers.Create, dr.Replicas[replicas.NodeID].Numbers.Destroy, dr.Replicas[replicas.NodeID].Numbers.Existing
}

func Default(replicasNumber uint64, existingIndexes []uint64, clusterSize uint64, member uint64) ([]uint64, []uint64, []uint64) {
	var create = make([]uint64, 0)
	var destroy = make([]uint64, 0)

	for replicas := 1 * member; replicas <= replicasNumber*member; replicas++ {
		create = append(create, replicas)
	}

	if len(create) < len(existingIndexes) {
		destroy = existingIndexes[len(create):]
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

	return create, destroy, existingIndexes
}

func (sr *ScopedReplicas) Add(group string, name string) {
	sr.Create = append(sr.Create, R{
		Group: group,
		Name:  name,
	})
}
func (sr *ScopedReplicas) Delete(group string, name string) {
	sr.Remove = append(sr.Remove, R{
		Group: group,
		Name:  name,
	})
}

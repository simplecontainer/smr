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

	dr.Replicas[replicas.NodeID].Numbers.Create, dr.Replicas[replicas.NodeID].Numbers.Destroy, dr.Replicas[replicas.NodeID].Numbers.Existing = replicas.GetReplicaNumbers(dr, containerDefinition.Spec.Container.Spread, replicas.Replicas, replicas.ExistingIndexes, len(clstr))

	create := dr.Replicas[replicas.NodeID].Numbers.Create
	destroy := dr.Replicas[replicas.NodeID].Numbers.Destroy

	// Destroy from the end to start
	if len(destroy) > 0 {
		for _, v := range destroy {
			name, _ := shared.Registry.NameReplicas(containerDefinition.Meta.Group, containerDefinition.Meta.Name, v)
			existingContainer := shared.Registry.Find(containerDefinition.Meta.Group, name)

			if existingContainer != nil {
				dr.Replicas[replicas.NodeID].Delete(existingContainer.GetGroup(), existingContainer.GetGeneratedName())
			}
		}
	}

	// Create from the start to the end
	for _, v := range create {
		name, _ := shared.Registry.NameReplicas(containerDefinition.Meta.Group, containerDefinition.Meta.Name, v)
		existingContainer := shared.Registry.Find(containerDefinition.Meta.Group, name)

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
		var err error

		containerObj, err = platforms.New(static.PLATFORM_DOCKER, name, shared.Manager.Config, containerDefinition)

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

	_, _, dr.Replicas[replicas.NodeID].Numbers.Existing = replicas.GetReplicaNumbers(dr, containerDefinition.Spec.Container.Spread, replicas.Replicas, replicas.ExistingIndexes, len(clstr))

	for i := range dr.Replicas[replicas.NodeID].Numbers.Existing {
		name, _ := shared.Registry.NameReplicas(containerDefinition.Meta.Group, containerDefinition.Meta.Name, i)
		containerObj := shared.Registry.Find(containerDefinition.Meta.Group, name)

		if containerObj != nil {
			containers = append(containers, R{replicas.Group, name})
		}
	}

	return containers, nil
}

func (replicas *Replicas) GetReplicaNumbers(dr *DistributedReplicas, spread v1.ContainerSpread, replicasNumber int, existingIndexes []int, clusterSize int) ([]int, []int, []int) {
	switch replicas.Spread.Spread {
	case platforms.SPREAD_SPECIFIC:
		for i := 1; i <= clusterSize; i++ {
			if dr.Replicas[uint64(i)] == nil {
				dr.Replicas[uint64(i)] = &ScopedReplicas{
					Create: make([]R, 0),
					Remove: make([]R, 0),
					Numbers: Numbers{
						Create:   make([]int, 0),
						Destroy:  make([]int, 0),
						Existing: make([]int, 0),
					},
				}
			}

			dr.Replicas[uint64(i)].Numbers.Create, dr.Replicas[uint64(i)].Numbers.Destroy, dr.Replicas[uint64(i)].Numbers.Existing = Specific(replicasNumber, existingIndexes, spread.Agents, i)
		}
		break
	case platforms.SPREAD_UNIFORM:
		for i := 1; i <= clusterSize; i++ {
			if dr.Replicas[uint64(i)] == nil {
				dr.Replicas[uint64(i)] = &ScopedReplicas{
					Create: make([]R, 0),
					Remove: make([]R, 0),
					Numbers: Numbers{
						Create:   make([]int, 0),
						Destroy:  make([]int, 0),
						Existing: make([]int, 0),
					},
				}
			}

			dr.Replicas[uint64(i)].Numbers.Create, dr.Replicas[uint64(i)].Numbers.Destroy, dr.Replicas[uint64(i)].Numbers.Existing = Uniform(replicasNumber, existingIndexes, clusterSize, i)
		}
		break
	default:
		dr.Replicas[replicas.NodeID].Numbers.Create, dr.Replicas[replicas.NodeID].Numbers.Destroy, dr.Replicas[replicas.NodeID].Numbers.Existing = Default(replicasNumber, existingIndexes, 1, 1)
		break
	}

	return dr.Replicas[replicas.NodeID].Numbers.Create, dr.Replicas[replicas.NodeID].Numbers.Destroy, dr.Replicas[replicas.NodeID].Numbers.Existing
}

func Default(replicasNumber int, existingIndexes []int, clusterSize int, member int) ([]int, []int, []int) {
	var create = make([]int, 0)
	var destroy = make([]int, 0)

	for replicas := 1 * member; replicas <= replicasNumber*member; replicas++ {
		create = append(create, replicas)
	}

	if len(create) < len(existingIndexes) {
		destroy = existingIndexes[len(create):]
	}

	return create, destroy, existingIndexes
}

func Uniform(replicasNumber int, existingIndexes []int, clusterSize int, member int) ([]int, []int, []int) {
	var create = make([]int, 0)
	var destroy = make([]int, 0)

	replicasScoped := replicasNumber / clusterSize

	for replicas := 1 * member; replicas <= replicasScoped*member; replicas++ {
		create = append(create, replicas)
	}

	if clusterSize == member {
		create = append(create, replicasNumber)
	}

	if len(create) < len(existingIndexes) {
		destroy = existingIndexes[len(create):]
	}

	return create, destroy, existingIndexes
}

func Specific(replicasNumber int, existingIndexes []int, nodes []int, member int) ([]int, []int, []int) {
	var create = make([]int, 0)
	var destroy = make([]int, 0)

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

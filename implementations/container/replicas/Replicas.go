package replicas

import (
	"errors"
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/implementations/container/container"
	"github.com/simplecontainer/smr/implementations/container/shared"
	"github.com/simplecontainer/smr/implementations/container/status"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
	"strings"
)

func (replicas *Replicas) HandleReplica(shared *shared.Shared, containerDefinition v1.Container, changelog diff.Changelog) ([]string, []string, error) {
	groups := make([]string, 0)
	names := make([]string, 0)

	numberOfReplicasToCreate, numberOfReplicasToDestroy, existingNumberOfReplicas := replicas.GetReplicaNumbers(replicas.Replicas, replicas.GeneratedIndex)

	// Destroy from the end to start
	if numberOfReplicasToDestroy > 0 {
		for i := existingNumberOfReplicas; i > (existingNumberOfReplicas - numberOfReplicasToDestroy); i -= 1 {
			name, _ := shared.Registry.NameReplicas(containerDefinition.Meta.Group, containerDefinition.Meta.Name, shared.Manager.Config.Environment.PROJECT, i)
			existingContainer := shared.Registry.Find(containerDefinition.Meta.Group, name)

			if existingContainer != nil {
				existingContainer.Status.TransitionState(existingContainer.Static.GeneratedName, status.STATUS_PENDING_DELETE)
			}

			groups = append(groups, replicas.Group)
			names = append(names, name)
		}
	}

	// Create from the start to the end
	for i := numberOfReplicasToCreate; i > 0; i -= 1 {
		name, _ := shared.Registry.NameReplicas(containerDefinition.Meta.Group, containerDefinition.Meta.Name, shared.Manager.Config.Environment.PROJECT, i)
		existingContainer := shared.Registry.Find(containerDefinition.Meta.Group, name)

		if existingContainer != nil {
			if existingContainer.Status.IfStateIs(status.STATUS_RECONCILING) {
				return nil, nil, errors.New("container is in reconciliation process try again later")
			}

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
				logger.Log.Debug("skipped recreating container since only scale up is triggered", zap.String("container", name), zap.String("group", replicas.Group))
				continue
			}
		}

		containerObj := container.NewContainerFromDefinition(shared.Manager.Config.Environment, name, containerDefinition)

		if existingContainer == nil {
			shared.Registry.AddOrUpdate(replicas.Group, name, shared.Manager.Config.Environment.PROJECT, containerObj)
			logger.Log.Debug("added container to registry", zap.String("container", name), zap.String("group", replicas.Group))
		} else {
			if replicas.Changed {
				shared.Registry.AddOrUpdate(replicas.Group, name, shared.Manager.Config.Environment.PROJECT, containerObj)
				logger.Log.Debug("update container since replica changed in registry", zap.String("container", name), zap.String("group", replicas.Group))
			}
		}

		groups = append(groups, replicas.Group)
		names = append(names, name)
	}

	return groups, names, nil
}

func (replicas *Replicas) GetReplica(shared *shared.Shared, containerDefinition v1.Container) ([]string, []string, error) {
	groups := make([]string, 0)
	names := make([]string, 0)

	_, _, existingNumberOfReplicas := replicas.GetReplicaNumbers(replicas.Replicas, replicas.GeneratedIndex)

	for i := existingNumberOfReplicas; i > 0; i -= 1 {
		name, _ := shared.Registry.NameReplicas(containerDefinition.Meta.Group, containerDefinition.Meta.Name, shared.Manager.Config.Environment.PROJECT, i)
		containerObj := shared.Registry.Find(containerDefinition.Meta.Group, name)

		if containerObj != nil {
			groups = append(groups, replicas.Group)
			names = append(names, name)
		}
	}

	return groups, names, nil
}

func (replicas *Replicas) GetReplicaNumbers(replicasNumber int, generatedNumber int) (int, int, int) {
	if replicasNumber > generatedNumber {
		return replicasNumber, 0, generatedNumber
	} else if replicasNumber == generatedNumber {
		return generatedNumber, 0, generatedNumber
	} else {
		return 0, generatedNumber - replicasNumber, generatedNumber
	}
}

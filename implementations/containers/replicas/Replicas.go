package replicas

import (
	"errors"
	"github.com/qdnqn/smr/implementations/container/container"
	"github.com/qdnqn/smr/implementations/container/status"
	"github.com/qdnqn/smr/implementations/containers/shared"
	"github.com/qdnqn/smr/pkg/database"
	"github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/r3labs/diff/v3"
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
			name, _ := shared.Registry.NameReplicas(containerDefinition.Meta.Group, containerDefinition.Meta.Name, shared.Manager.Runtime.PROJECT, i)
			existingContainer := shared.Registry.Find(containerDefinition.Meta.Group, name)

			if existingContainer != nil {
				existingContainer.Status.TransitionState(status.STATUS_PENDING_DELETE)
			}

			groups = append(groups, replicas.Group)
			names = append(names, name)
		}
	}

	// Create from the start to the end
	for i := numberOfReplicasToCreate; i > 0; i -= 1 {
		name, _ := shared.Registry.NameReplicas(containerDefinition.Meta.Group, containerDefinition.Meta.Name, shared.Manager.Runtime.PROJECT, i)
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
				logger.Log.Info("skipped recreating container since only scale up is triggered", zap.String("container", name), zap.String("group", replicas.Group))
				continue
			}
		}

		containerObj := container.NewContainerFromDefinition(shared.Manager.Runtime, name, containerDefinition)

		for i, v := range containerObj.Runtime.Resources {
			format := database.Format("resource", containerObj.Static.Group, v.Identifier, v.Key)
			val, err := database.Get(shared.Manager.Badger, format.ToString())

			if err != nil {
				logger.Log.Error("failed to get resources for the container")
			}

			containerObj.Runtime.Resources[i].Data[v.Key] = val
		}

		logger.Log.Info("retrieved resources for container", zap.String("container", name))

		if existingContainer == nil {
			shared.Registry.AddOrUpdate(replicas.Group, name, shared.Manager.Runtime.PROJECT, containerObj)
			logger.Log.Info("added container to registry", zap.String("container", name), zap.String("group", replicas.Group))
		} else {
			if replicas.Changed {
				shared.Registry.AddOrUpdate(replicas.Group, name, shared.Manager.Runtime.PROJECT, containerObj)
				logger.Log.Info("update container since replica changed in registry", zap.String("container", name), zap.String("group", replicas.Group))
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
		name, _ := shared.Registry.NameReplicas(containerDefinition.Meta.Group, containerDefinition.Meta.Name, shared.Manager.Runtime.PROJECT, i)
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

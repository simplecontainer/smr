package replicas

import (
	"errors"
	"github.com/qdnqn/smr/pkg/container"
	"github.com/qdnqn/smr/pkg/database"
	"github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/qdnqn/smr/pkg/manager"
	"github.com/r3labs/diff/v3"
	"go.uber.org/zap"
	"strings"
)

func (replicas *Replicas) HandleContainer(mgr *manager.Manager, containerDefinition v1.Container, changelog diff.Changelog) ([]string, []string, error) {
	groups := make([]string, 0)
	names := make([]string, 0)

	numberOfReplicasToCreate, numberOfReplicasToDestroy, existingNumberOfReplicas := replicas.GetReplicaNumbers(replicas.Replicas, replicas.GeneratedIndex)

	if numberOfReplicasToDestroy > 0 {
		for i := existingNumberOfReplicas; i > (existingNumberOfReplicas - numberOfReplicasToDestroy); i -= 1 {
			name := containerDefinition.Meta.Name
			container := container.NewContainerFromDefinition(mgr.Runtime, name, containerDefinition)

			existingContainer := mgr.Registry.Find(container.Static.Group, name)

			if existingContainer != nil {
				existingContainer.Status.PendingDelete = true
			}

			groups = append(groups, replicas.Group)
			names = append(names, name)
		}
	}

	for i := numberOfReplicasToCreate; i > 0; i -= 1 {
		name := containerDefinition.Meta.Name
		container := container.NewContainerFromDefinition(mgr.Runtime, name, containerDefinition)

		for i, v := range container.Runtime.Resources {
			format := database.Format("resource", container.Static.Group, v.Identifier, v.Key)
			val, err := database.Get(mgr.Badger, format.ToString())

			if err != nil {
				logger.Log.Error("failed to get resources for the container")
			}

			container.Runtime.Resources[i].Data[v.Key] = val
		}

		logger.Log.Info("retrieved resources for container", zap.String("container", name))

		/*
			Do all pre-checks here before rewriting container in the registry
		*/

		logger.Log.Info("checking if pre-check conditions ready before add/update container in registry", zap.String("container", name))
		existingContainer := mgr.Registry.Find(container.Static.Group, name)

		if existingContainer != nil {
			logger.Log.Info("container already existing on the server", zap.String("container", name))

			if existingContainer.Status.Reconciling {
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

			// If container got to here without any failures we need to set it definitionDrift=true so that we do reconcile
			// in the container implementation
			container.Status.DefinitionDrift = true
		}

		mgr.Registry.AddOrUpdate(replicas.Group, name, mgr.Runtime.PROJECT, container)
		logger.Log.Info("added container to registry", zap.String("container", name), zap.String("group", replicas.Group))

		groups = append(groups, replicas.Group)
		names = append(names, name)
	}

	return groups, names, nil
}

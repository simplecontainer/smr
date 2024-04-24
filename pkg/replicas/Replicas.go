package replicas

import (
	"errors"
	"go.uber.org/zap"
	"smr/pkg/container"
	"smr/pkg/database"
	"smr/pkg/definitions"
	"smr/pkg/logger"
	"smr/pkg/manager"
)

func (replicas *Replicas) HandleReplica(mgr *manager.Manager, containerDefinition definitions.Container) ([]string, []string, error) {
	groups := make([]string, 0)
	names := make([]string, 0)

	numberOfReplicas := 0

	if replicas.GeneratedIndex > containerDefinition.Spec.Container.Replicas {
		numberOfReplicas = replicas.GeneratedIndex - 1
	} else {
		numberOfReplicas = containerDefinition.Spec.Container.Replicas
	}

	for i := numberOfReplicas; i > 0; i -= 1 {
		name, _ := mgr.Registry.NameReplicas(containerDefinition.Meta.Group, containerDefinition.Meta.Name, mgr.Runtime.PROJECT, i)
		container := container.NewContainerFromDefinition(mgr.Runtime, name, containerDefinition)

		for i, v := range container.Runtime.Resources {
			format := database.Format("resource", container.Static.Group, v.Identifier, v.Key)
			val, err := database.Get(mgr.Badger, format.ToString())

			if err != nil {
				logger.Log.Error("Failed to get resources for the container")
			}

			container.Runtime.Resources[i].Data[v.Key] = val
		}

		logger.Log.Info("retrieved resources for container", zap.String("container", container.Static.Name))

		/*
			Do all pre-checks here before rewriting container in the registry
		*/

		logger.Log.Info("checking if pre-check conditions ready before add/update container in registry", zap.String("container", container.Static.Name))
		existingContainer := mgr.Registry.Find(container.Static.Group, name)

		if existingContainer != nil {
			logger.Log.Info("container already existing on the server", zap.String("container", container.Static.Name))

			if existingContainer.Status.Reconciling {
				return nil, nil, errors.New("container is in reconciliation process try again later")
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

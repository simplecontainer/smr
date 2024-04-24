package replicas

import (
	"fmt"
	"go.uber.org/zap"
	"smr/pkg/container"
	"smr/pkg/database"
	"smr/pkg/definitions"
	"smr/pkg/logger"
	"smr/pkg/manager"
	"strings"
)

func (replicas *Replicas) HandleReplica(mgr *manager.Manager, containerDefinition definitions.Definition) ([]string, []string) {
	groups := make([]string, 0)
	names := make([]string, 0)

	numberOfReplicas := 0

	if replicas.GeneratedIndex > containerDefinition.Spec.Container.Replicas {
		numberOfReplicas = replicas.GeneratedIndex - 1
	} else {
		numberOfReplicas = containerDefinition.Spec.Container.Replicas
	}

	for i := numberOfReplicas; i > 0; i -= 1 {
		name, _ := mgr.Registry.NameReplicas(replicas.Group, mgr.Runtime.PROJECT, i)
		container := container.NewContainerFromDefinition(mgr.Runtime, name, containerDefinition)

		for i, v := range container.Runtime.Resources {
			format := database.Format("resource", container.Static.Group, v.Identifier, v.Key)
			key := strings.TrimSpace(fmt.Sprintf("%s.%s.%s.%s", format.Kind, format.Group, format.Identifier, format.Key))
			val, err := database.Get(mgr.Badger, key)

			if err != nil {
				logger.Log.Error("Failed to get resources for the container")
			}

			container.Runtime.Resources[i].Data[v.Key] = val
		}

		logger.Log.Info("retrieved resources for container", zap.String("container", container.Static.Name))

		if container.Get() != nil {
			if mgr.Registry.Containers[container.Static.Group][container.Static.Name] != nil {
				logger.Log.Info("Checking if change in definition")

				if definitions.Compare(mgr.Registry.Containers[container.Static.Group][container.Static.Name].Static.Definition, container.Static.Definition) {
					logger.Log.Info("definition is same, ignoring request")
					continue
				}
			} else {
				container.Stop()
				container.Delete()
				mgr.Registry.Remove(replicas.Group, name, mgr.Runtime.PROJECT)
			}
		}

		mgr.Registry.AddOrUpdate(replicas.Group, name, mgr.Runtime.PROJECT, container)
		logger.Log.Info("added container to registry", zap.String("container", container.Static.Name))

		groups = append(groups, replicas.Group)
		names = append(names, name)
	}

	return groups, names
}

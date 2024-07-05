package replicas

import (
	"errors"
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/implementations/container/container"
	"github.com/simplecontainer/smr/implementations/container/shared"
	"github.com/simplecontainer/smr/implementations/container/status"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/objects"
	"go.uber.org/zap"
	"strings"
)

func (replicas *Replicas) HandleContainer(shared *shared.Shared, mgr *manager.Manager, containerDefinition v1.Container, changelog diff.Changelog) ([]string, []string, error) {
	groups := make([]string, 0)
	names := make([]string, 0)

	numberOfReplicasToCreate, numberOfReplicasToDestroy, existingNumberOfReplicas := replicas.GetReplicaNumbers(replicas.Replicas, replicas.GeneratedIndex)

	if numberOfReplicasToDestroy > 0 {
		for i := existingNumberOfReplicas; i > (existingNumberOfReplicas - numberOfReplicasToDestroy); i -= 1 {
			name := containerDefinition.Meta.Name
			container := container.NewContainerFromDefinition(mgr.Config.Environment, name, containerDefinition)

			existingContainer := shared.Registry.Find(container.Static.Group, name)

			if existingContainer != nil {
				existingContainer.Status.TransitionState(existingContainer.Static.GeneratedName, status.STATUS_PENDING_DELETE)
			}

			groups = append(groups, replicas.Group)
			names = append(names, name)
		}
	}

	for i := numberOfReplicasToCreate; i > 0; i -= 1 {
		name := containerDefinition.Meta.Name
		container := container.NewContainerFromDefinition(mgr.Config.Environment, name, containerDefinition)

		for i, v := range container.Runtime.Resources {
			format := objects.Format("resource", container.Static.Group, v.Identifier, v.Key)

			obj := objects.New()
			err := obj.Find(shared.Client, format)

			if err != nil {
				logger.Log.Error("failed to get resources for the container")
			}

			container.Runtime.Resources[i].Data[v.Key] = obj.GetDefinitionString()
		}

		logger.Log.Info("retrieved resources for container", zap.String("container", name))

		/*
			Do all pre-checks here before rewriting container in the registry
		*/

		logger.Log.Info("checking if pre-check conditions ready before add/update container in registry", zap.String("container", name))
		existingContainer := shared.Registry.Find(container.Static.Group, name)

		if existingContainer != nil {
			logger.Log.Info("container already existing on the server", zap.String("container", name))

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

		shared.Registry.AddOrUpdate(replicas.Group, name, mgr.Config.Environment.PROJECT, container)
		logger.Log.Info("added container to registry", zap.String("container", name), zap.String("group", replicas.Group))

		groups = append(groups, replicas.Group)
		names = append(names, name)
	}

	return groups, names, nil
}

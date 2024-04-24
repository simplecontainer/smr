package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/r3labs/diff/v3"
	"go.uber.org/zap"
	"smr/pkg/database"
	"smr/pkg/definitions"
	"smr/pkg/dependency"
	"smr/pkg/implementations"
	"smr/pkg/logger"
	"smr/pkg/manager"
	"smr/pkg/objects"
	"smr/pkg/reconciler"
	"smr/pkg/replicas"
)

func (implementation *Implementation) Implementation(mgr *manager.Manager, jsonData []byte) (implementations.Response, error) {
	definitionSent := &definitions.Containers{}

	if err := json.Unmarshal(jsonData, &definitionSent); err != nil {
		return implementations.Response{
			HttpStatus:       400,
			Explanation:      "invalid definition sent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}, err
	} else {
		var globalGroups []string
		var globalNames []string

		for _, definition := range definitionSent.Containers {
			format := database.Format("container", definition.Meta.Group, definition.Meta.Name, "object")
			obj := objects.New()
			err = obj.Find(mgr.Registry.Object, mgr.Badger, format)

			var jsonStringFromRequest string
			jsonStringFromRequest, err = definition.ToJsonString()

			if obj.Exists() {
				if obj.Diff(jsonStringFromRequest) {
					// Detect only change on replicas, if that's true tackle only scale up or scale down without recreating
					// containers that are there already, otherwise recreate everything
				}
			} else {
				err = obj.Add(mgr.Registry.Object, mgr.Badger, format, jsonStringFromRequest)
			}

			if obj.ChangeDetected() || !obj.Exists() {
				logger.Log.Info("object is changed", zap.String("container", definition.Meta.Name))

				name := definition.Meta.Name
				logger.Log.Info(fmt.Sprintf("trying to generate container %s object", name))

				_, ok := definitionSent.Containers[name]

				if !ok {
					return implementations.Response{
						HttpStatus:       400,
						Explanation:      "container definition invalid",
						ErrorExplanation: fmt.Sprintf("container definintion with name %s not found", name),
						Error:            true,
						Success:          false,
					}, errors.New(fmt.Sprintf("container definintion with name %s not found", name))
				}

				groups, names, err := implementation.generateReplicaNamesAndGroups(mgr, definitionSent.Containers[name], obj.Changelog)

				if err == nil {
					logger.Log.Info(fmt.Sprintf("generated container %s object", name))

					globalGroups = append(globalGroups, groups...)
					globalNames = append(globalNames, names...)

					err = obj.Update(mgr.Registry.Object, mgr.Badger, format, jsonStringFromRequest)
				} else {
					logger.Log.Error("failed to generate names and groups")

					return implementations.Response{
						HttpStatus:       500,
						Explanation:      "failed to generate groups and names",
						ErrorExplanation: err.Error(),
						Error:            false,
						Success:          true,
					}, err
				}
			} else {
				logger.Log.Info("object is same on the server", zap.String("container", definition.Meta.Name))
			}
		}

		if len(globalGroups) > 0 {
			/*
			   Order contains pool of containers from registry.
			   It is possible to read all information if container is already in registry.
			   With generated names and groups orderByDependencies will order them by dependencies.
			   Registry is already pre-populated by handleReplicas if container is not found by name and group in registry.

			   All containers existing in order should be reconciled
			*/

			logger.Log.Info(fmt.Sprintf("trying to order containers by dependencies"))
			order := implementation.orderByDependencies(mgr.Registry, globalGroups, globalNames)
			logger.Log.Info(fmt.Sprintf("containers are ordered by dependencies"))

			var solved bool

			for _, container := range order {
				if container.Status.PendingDelete {
					logger.Log.Info(fmt.Sprintf("container is pending to delete %s", container.Static.GeneratedName))

					mgr.Registry.Remove(container.Static.Group, container.Static.GeneratedName)

					mgr.Reconciler.QueueChan <- reconciler.Reconcile{
						Container: container,
					}
				} else {
					solved, err = dependency.Ready(mgr, container.Static.Group, container.Static.GeneratedName, container.Static.Definition.Spec.Container.Dependencies)

					if solved {
						if container.Status.DefinitionDrift {
							// This the case when we know container already exists and definition was reapplied
							// We should trigger the reconcile
							logger.Log.Info("sending container to reconcile state", zap.String("container", container.Static.GeneratedName))

							mgr.Reconciler.QueueChan <- reconciler.Reconcile{
								Container: container,
							}
						} else {
							// This the case when we know container doesn't exist, and we are running it the first time

							logger.Log.Info("trying to run container", zap.String("group", container.Static.Group), zap.String("name", container.Static.Name))

							container.Prepare(mgr.Badger)
							_, err = container.Run(mgr.Runtime, mgr.Badger, mgr.DnsCache)

							if err != nil {
								return implementations.Response{
									HttpStatus:       500,
									Explanation:      "failed to start container",
									ErrorExplanation: err.Error(),
									Error:            true,
									Success:          false,
								}, err
							}
						}
					} else {
						return implementations.Response{
							HttpStatus:       500,
							Explanation:      "failed to solve container dependencies",
							ErrorExplanation: err.Error(),
							Error:            true,
							Success:          false,
						}, err
					}
				}
			}

			return implementations.Response{
				HttpStatus:       200,
				Explanation:      "everything went smoothly: good job!",
				ErrorExplanation: "",
				Error:            false,
				Success:          true,
			}, nil
		} else {
			return implementations.Response{
				HttpStatus:       200,
				Explanation:      "object is same as the one on the server",
				ErrorExplanation: "",
				Error:            false,
				Success:          true,
			}, err
		}
	}
}

func (implementation *Implementation) generateReplicaNamesAndGroups(mgr *manager.Manager, containerDefinition definitions.Container, changelog diff.Changelog) ([]string, []string, error) {
	_, index := mgr.Registry.Name(containerDefinition.Meta.Group, containerDefinition.Meta.Name, mgr.Runtime.PROJECT)

	r := replicas.Replicas{
		Group:          containerDefinition.Meta.Group,
		GeneratedIndex: index - 1,
		Replicas:       containerDefinition.Spec.Container.Replicas,
	}

	groups, names, err := r.HandleReplica(mgr, containerDefinition, changelog)

	return groups, names, err
}

// Exported
var Containers Implementation

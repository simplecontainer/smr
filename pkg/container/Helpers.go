package container

import (
	"context"
	"encoding/json"
	"github.com/dgraph-io/badger/v4"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	v1 "github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/logger"
	"github.com/qdnqn/smr/pkg/network"
	"github.com/qdnqn/smr/pkg/runtime"
	"log"
	"os"
	"time"
)

func (container *Container) mappingToMounts(BadgerEncrypted *badger.DB, runtime *runtime.Runtime) []mount.Mount {
	var mounts []mount.Mount

	for _, v := range container.Runtime.Resources {
		tmpFile, err := os.CreateTemp("/tmp", container.Static.Name)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := tmpFile.WriteString(container.UnpackSecretsResources(BadgerEncrypted, v.Data[v.Key].(string))); err != nil {
			log.Fatal(err)
		}

		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: tmpFile.Name(),
			Target: v.MountPoint,
		})
	}

	for _, v := range container.Static.MappingFiles {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: v["host"],
			Target: v["target"],
		})
	}

	return mounts
}

func (container *Container) portMappings() map[nat.Port][]nat.PortBinding {
	var portBindings = make(map[nat.Port][]nat.PortBinding, len(container.Static.MappingPorts))

	for _, v := range container.Static.MappingPorts {
		portBindings[nat.Port(v.Container)] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: v.Host,
			},
		}
	}

	return portBindings
}
func (container *Container) exposedPorts() map[nat.Port]struct{} {
	var exposedPorts = make(map[nat.Port]struct{}, len(container.Static.ExposedPorts))

	for _, v := range container.Static.ExposedPorts {
		exposedPorts[nat.Port(v)] = struct{}{}
	}

	return exposedPorts
}

func mapAnyToResources(definition []map[string]any) []Resource {
	mapToJson, err := json.Marshal(definition)
	if err != nil {
		logger.Log.Error(err.Error())
		return []Resource{}
	}

	var resources []Resource
	if err := json.Unmarshal(mapToJson, &resources); err != nil {
		logger.Log.Error(err.Error())
		return []Resource{}
	}

	for i, _ := range resources {
		resources[i].Data = make(map[string]any)
	}

	return resources
}

func convertPortMappingsToExposedPorts(portMappings []network.PortMappings) []string {
	var exposedPorts []string

	for _, portMap := range portMappings {
		exposedPorts = append(exposedPorts, portMap.Container)
	}

	return exposedPorts
}

func convertReadinessDefinitionToReadiness(readinessDefinition []v1.Readiness) []Readiness {
	var readiness = make([]Readiness, 0)

	for _, val := range readinessDefinition {
		if val.Timeout == "" {
			val.Timeout = "30s"
		}

		timeout, err := time.ParseDuration(val.Timeout)

		var ctx context.Context
		if err == nil {
			ctx, _ = context.WithTimeout(context.Background(), timeout)
		} else {
			return nil
		}

		readinessTmp := Readiness{
			Name:     val.Name,
			Operator: val.Operator,
			Timeout:  val.Timeout,
			Body:     val.Body,
			Solved:   false,
			Ctx:      ctx,
		}

		readiness = append(readiness, readinessTmp)
	}

	return readiness
}

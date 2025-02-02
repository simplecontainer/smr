package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/smaps"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/simplecontainer/smr/pkg/template"
	"os"
)

func (container *Docker) PrepareConfiguration(client *client.Http, user *authentication.User, runtime *types.Runtime) error {
	var parsed string
	var dependencies []f.Format
	var err error

	container.Configuration.Map.Range(func(key, value any) bool {
		parsed, dependencies, err = template.Parse(key.(string), value.(string), client, user, nil)

		if err != nil {
			return false
		}

		runtime.Configuration.Map.Store(key, parsed)
		runtime.ObjectDependencies = append(runtime.ObjectDependencies, dependencies...)

		return true
	})

	return err
}

func (container *Docker) PrepareResources(client *client.Http, user *authentication.User, runtime *types.Runtime) error {
	err := container.Volumes.RemoveResources()
	if err != nil {
		return err
	}

	for k, v := range container.Resources.Resources {
		format := f.New(static.SMR_PREFIX, "kind", static.KIND_RESOURCE, v.Reference.Group, v.Reference.Name)

		obj := objects.New(client.Get(user.Username), user)
		err = obj.Find(format)

		if !obj.Exists() {
			return errors.New(fmt.Sprintf("failed to fetch resource from the kv store %s", format.ToString()))
		}

		resourceObj := v1.ResourceDefinition{}

		err = json.Unmarshal(obj.GetDefinitionByte(), &resourceObj)

		if err != nil {
			return err
		}

		container.Resources.Resources[k].Docker.Data = smaps.NewFromMap(resourceObj.Spec.Data)

		container.Resources.Resources[k].Docker.Data.Map.Range(func(key, value any) bool {
			var parsed string
			parsed, _, err = template.Parse(key.(string), value.(string), client, user, runtime.Configuration)

			container.Resources.Resources[k].Docker.Data.Map.Store(key, parsed)

			if err != nil {
				return false
			}

			return true
		})

		if err != nil {
			return err
		}

		var tmpFile *os.File
		tmpFile, err = os.CreateTemp("/tmp", container.Name)

		if err != nil {
			return err
		}

		val, ok := container.Resources.Resources[k].Docker.Data.Map.Load(v.Reference.Key)

		if !ok {
			return errors.New(fmt.Sprintf("key %s doesnt exist in resource %s", v.Reference.Key, v.Reference.Name))
		}

		if _, err = tmpFile.WriteString(val.(string)); err != nil {
			return err
		}

		err = container.Volumes.Add(v1.ContainerVolume{
			Type:       "resource",
			HostPath:   tmpFile.Name(),
			MountPoint: v.Reference.MountPoint,
		})

		if err != nil {
			return err
		}

		runtime.ObjectDependencies = append(runtime.ObjectDependencies, f.New(static.SMR_PREFIX, static.CATEGORY_KIND, "resource", container.Resources.Resources[k].Reference.Group, container.Resources.Resources[k].Reference.Name))
	}

	return nil
}

func (container *Docker) PrepareLabels(runtime *types.Runtime) error {
	var parsed string
	var err error

	container.Labels.Map.Range(func(key, value any) bool {
		parsed, _, err = template.Parse(key.(string), value.(string), nil, nil, runtime.Configuration)

		if err != nil {
			return false
		}

		container.Labels.Map.Store(key.(string), parsed)
		return true
	})

	return err
}

func (container *Docker) PrepareEnvs(runtime *types.Runtime) error {
	var err error

	for index, value := range container.Env {
		container.Env[index], _, err = template.Parse(value, value, nil, nil, runtime.Configuration)

		if err != nil {
			return err
		}
	}

	return nil
}

func (container *Docker) PrepareReadiness(runtime *types.Runtime) error {
	var err error

	for indexReadiness, _ := range container.Readiness.Readinesses {
		for index, _ := range container.Readiness.Readinesses[indexReadiness].Reference.Data {
			container.Lock.Lock()
			container.Readiness.Readinesses[indexReadiness].Docker.Body[index], _, err = template.Parse(index, container.Readiness.Readinesses[indexReadiness].Docker.Body[index], nil, nil, runtime.Configuration)
			container.Lock.Unlock()

			if err != nil {
				return err
			}
		}
	}

	return nil
}

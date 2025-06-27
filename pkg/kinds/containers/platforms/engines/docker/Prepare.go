package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/configuration"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/types"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/smaps"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/simplecontainer/smr/pkg/template"
	"os"
	"strconv"
)

func (container *Docker) PrepareConfiguration(config *configuration.Configuration, client *clients.Http, user *authentication.User, runtime *types.Runtime) error {
	var parsed string
	var dependencies []f.Format
	var err error
	var index uint64

	index, err = container.GetIndex()

	if err != nil {
		return err
	}

	runtime.Configuration.Map.Store("name", container.GetGeneratedName())
	runtime.Configuration.Map.Store("node", config.NodeName)
	runtime.Configuration.Map.Store("index", strconv.FormatUint(index, 10))
	runtime.Configuration.Map.Store("fqdn", container.GetDomain(static.CLUSTER_NETWORK))
	runtime.Configuration.Map.Store("headless", container.GetHeadlessDomain(static.CLUSTER_NETWORK))

	container.Configuration.Map.Range(func(key, value any) bool {
		parsed, dependencies, err = template.Parse(key.(string), value.(string), client, user, runtime.Configuration, 0)

		if err != nil {
			return false
		}

		runtime.Configuration.Map.Store(key, parsed)
		runtime.ObjectDependencies = append(runtime.ObjectDependencies, dependencies...)

		return true
	})

	return err
}

func (container *Docker) PrepareConfigurations(client *clients.Http, user *authentication.User, runtime *types.Runtime) error {
	err := container.Volumes.RemoveResources()
	if err != nil {
		return err
	}

	for k, v := range container.Configurations.Configurations {
		format := f.New(static.SMR_PREFIX, "kind", static.KIND_CONFIGURATION, v.Reference.Group, v.Reference.Name)

		obj := objects.New(client.Get(user.Username), user)
		err = obj.Find(format)

		if !obj.Exists() {
			return errors.New(fmt.Sprintf("failed to fetch resource from the kv store %s", format.ToString()))
		}

		configurationObj := v1.ConfigurationDefinition{}

		err = json.Unmarshal(obj.GetDefinitionByte(), &configurationObj)

		if err != nil {
			return err
		}

		for key, value := range configurationObj.Spec.Data {
			var parsed string
			parsed, _, err = template.Parse(key, value, client, user, runtime.Configuration, 0)

			if err != nil {
				return err
			}

			runtime.Configuration.Add(key, parsed)
		}

		runtime.ObjectDependencies = append(runtime.ObjectDependencies, f.New(static.SMR_PREFIX, static.CATEGORY_KIND, static.KIND_CONFIGURATION, container.Resources.Resources[k].Reference.Group, container.Resources.Resources[k].Reference.Name))
	}

	return nil
}

func (container *Docker) PrepareResources(client *clients.Http, user *authentication.User, runtime *types.Runtime) error {
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
			parsed, _, err = template.Parse(key.(string), value.(string), client, user, runtime.Configuration, 0)

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

		err = tmpFile.Chmod(0666)

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

		err = container.Volumes.Add(container.GetGeneratedName(), v1.ContainersVolume{
			Type:       "resource",
			HostPath:   tmpFile.Name(),
			MountPoint: v.Reference.MountPoint,
		})

		if err != nil {
			return err
		}

		runtime.ObjectDependencies = append(runtime.ObjectDependencies, f.New(static.SMR_PREFIX, static.CATEGORY_KIND, static.KIND_RESOURCE, container.Resources.Resources[k].Reference.Group, container.Resources.Resources[k].Reference.Name))
	}

	return nil
}

func (container *Docker) PrepareLabels(runtime *types.Runtime) error {
	var parsed string
	var err error

	container.Labels.Labels.Map.Range(func(key, value any) bool {
		parsed, _, err = template.Parse(key.(string), value.(string), nil, nil, runtime.Configuration, 0)

		if err != nil {
			return false
		}

		container.Labels.Labels.Map.Store(key.(string), parsed)
		return true
	})

	return err
}

func (container *Docker) PrepareEnvs(runtime *types.Runtime) error {
	var err error

	for index, value := range container.Env {
		container.Env[index], _, err = template.Parse(value, value, nil, nil, runtime.Configuration, 0)

		if err != nil {
			return err
		}
	}

	return nil
}

func (container *Docker) PrepareAuth(runtime *types.Runtime) error {
	var err error

	container.RegistryAuth, _, err = template.Parse(container.RegistryAuth, container.RegistryAuth, nil, nil, runtime.Configuration, 0)

	if err != nil {
		return err
	}

	return nil
}

func (container *Docker) PrepareReadiness(runtime *types.Runtime) error {
	var err error

	for indexReadiness, _ := range container.Readiness.Readinesses {
		for index, val := range container.Readiness.Readinesses[indexReadiness].Body {
			container.Lock.Lock()
			container.Readiness.Readinesses[indexReadiness].BodyUnpack[index], _, err = template.Parse(index, val, nil, nil, runtime.Configuration, 0)
			container.Lock.Unlock()

			if err != nil {
				return err
			}
		}

		container.Readiness.Readinesses[indexReadiness].CommandUnpacked = make([]string, len(container.Readiness.Readinesses[indexReadiness].Command))

		container.Readiness.Readinesses[indexReadiness].URL, _, err = template.Parse("readiness-url", container.Readiness.Readinesses[indexReadiness].URL, nil, nil, runtime.Configuration, 0)

		if err != nil {
			return err
		}

		for index, val := range container.Readiness.Readinesses[indexReadiness].Command {
			container.Lock.Lock()
			container.Readiness.Readinesses[indexReadiness].CommandUnpacked[index], _, err = template.Parse("readiness-command", val, nil, nil, runtime.Configuration, 0)
			container.Lock.Unlock()

			if err != nil {
				return err
			}
		}
	}

	return nil
}

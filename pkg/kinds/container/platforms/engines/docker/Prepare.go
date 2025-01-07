package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/secrets"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/smaps"
	"github.com/simplecontainer/smr/pkg/template"
	"os"
	"regexp"
	"strings"
)

func (container *Docker) PrepareNetwork(client *client.Http, user *authentication.User, runtime *types.Runtime) error {
	for _, network := range container.Networks.Networks {
		runtime.Configuration.Map.Store(fmt.Sprintf("%s_hostname", network.Reference.Name), container.GetDomain(network.Reference.Name))
	}

	return nil
}

func (container *Docker) PrepareConfiguration(client *client.Http, user *authentication.User, runtime *types.Runtime) error {
	var dependencyMap []*f.Format
	var err error

	obj := objects.New(client.Get(user.Username), user)

	// Scan configuration and parse all placeholders and append it to the runtime so that rest of the gang can use it
	container.Configuration.Map.Range(func(key, value any) bool {
		var parsed string
		parsed, runtime.ObjectDependencies, err = template.ParseTemplate(obj, value.(string), nil)

		if err != nil {
			return false
		}

		runtime.Configuration.Map.Store(key, parsed)

		return true
	})

	if err != nil {
		return err
	}

	runtime.ObjectDependencies = append(runtime.ObjectDependencies, dependencyMap...)
	return nil
}

func (container *Docker) PrepareResources(client *client.Http, user *authentication.User, runtime *types.Runtime) error {
	err := container.Volumes.RemoveResources()
	if err != nil {
		return err
	}

	for k, v := range container.Resources.Resources {
		format := f.New("resource", v.Reference.Group, v.Reference.Name, "object")

		obj := objects.New(client.Get(user.Username), user)
		err = obj.Find(format)

		if !obj.Exists() {
			return errors.New(fmt.Sprintf("failed to fetch resource from the kv store %s", format.ToString()))
		}

		resourceObject := v1.ResourceDefinition{}

		err = json.Unmarshal(obj.GetDefinitionByte(), &resourceObject)

		if err != nil {
			return err
		}

		container.Resources.Resources[k].Docker.Data = smaps.NewFromMap(resourceObject.Spec.Data)

		container.Resources.Resources[k].Docker.Data.Map.Range(func(key, value any) bool {
			var parsed string
			parsed, _, err = template.ParseTemplate(obj, value.(string), runtime.Configuration)

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

		var resource string
		resource, err = secrets.UnpackSecretsResources(client, user, val.(string))

		if err != nil {
			return err
		}

		if _, err = tmpFile.WriteString(resource); err != nil {
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

		runtime.ObjectDependencies = append(runtime.ObjectDependencies, &f.Format{
			Kind:       "resource",
			Group:      container.Resources.Resources[k].Reference.Group,
			Identifier: container.Resources.Resources[k].Reference.Name,
			Key:        "",
		})
	}

	return nil
}

func (container *Docker) PrepareLabels(runtime *types.Runtime) {
	container.Labels.Map.Range(func(key, value any) bool {
		regexDetectBigBrackets := regexp.MustCompile(`{{([^{\n}]*)}}`)
		matches := regexDetectBigBrackets.FindAllStringSubmatch(key.(string), -1)

		if len(matches) > 0 {
			trimmedMatch := strings.TrimSpace(matches[0][1])
			SplitByDot := strings.SplitN(trimmedMatch, ".", 2)

			runtimeValue, ok := runtime.Configuration.Map.Load(SplitByDot[1])

			if len(SplitByDot) > 1 && ok && runtimeValue != "" {
				newIndex := strings.Replace(key.(string), matches[0][0], runtimeValue.(string), 1)

				container.Labels.Map.Store(newIndex, value)
				container.Labels.Map.Delete(key)
			}
		}

		return true
	})
}

func (container *Docker) PrepareEnvs(runtime *types.Runtime) {
	for index, value := range container.Env {
		regexDetectBigBrackets := regexp.MustCompile(`{{([^{\n}]*)}}`)
		matches := regexDetectBigBrackets.FindAllStringSubmatch(value, -1)

		if len(matches) > 0 {
			SplitByDot := strings.SplitN(matches[0][1], ".", 2)

			trimmedIndex := strings.TrimSpace(SplitByDot[1])

			runtimeValue, ok := runtime.Configuration.Map.Load(trimmedIndex)

			if len(SplitByDot) > 1 && ok && runtimeValue != "" {
				container.Env[index] = strings.Replace(container.Env[index], matches[0][0], runtimeValue.(string), 1)
			}
		}
	}
}

func (container *Docker) PrepareReadiness(runtime *types.Runtime) {
	for indexReadiness, _ := range container.Readiness.Readinesses {
		for index, value := range container.Readiness.Readinesses[indexReadiness].Reference.Data {
			regexDetectBigBrackets := regexp.MustCompile(`{{([^{\n}]*)}}`)
			matches := regexDetectBigBrackets.FindAllStringSubmatch(value, -1)

			if len(matches) > 0 {
				format := f.NewFromString(matches[0][1])

				if format.IsValid() && format.Kind == "secret" {
					continue
				} else {
					container.Lock.Lock()
					runtimeValue, _ := runtime.Configuration.Map.Load(format.Group)

					container.Readiness.Readinesses[indexReadiness].Docker.Body[index] = strings.Replace(container.Readiness.Readinesses[indexReadiness].Docker.Body[index], matches[0][0], runtimeValue.(string), 1)
					container.Lock.Unlock()
				}
			}
		}
	}
}

package container

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/template"
	"log"
	"os"
	"regexp"
	"strings"
)

func (container *Container) Prepare(client *client.Http, user *authentication.User) error {
	err := container.PrepareNetwork(client, user)

	if err != nil {
		return err
	}

	err = container.PrepareConfiguration(client, user)

	if err != nil {
		return err
	}

	err = container.PrepareResources(client, user)

	if err != nil {
		return err
	}

	container.PrepareLabels()
	container.PrepareEnvs()
	container.PrepareReadiness()

	return nil
}

func (container *Container) PrepareNetwork(client *client.Http, user *authentication.User) error {
	for _, network := range container.Static.Networks.Networks {
		container.Runtime.Configuration[fmt.Sprintf("%s_hostname", network)] = container.GetDomain(network.Reference.Name)
	}

	return nil
}

func (container *Container) PrepareConfiguration(client *client.Http, user *authentication.User) error {
	var dependencyMap []*f.Format
	var err error

	format := f.New("configuration", container.Static.Group, container.Static.GeneratedName, "")

	obj := objects.New(client.Get(user.Username), user)
	container.Runtime.Configuration, container.Runtime.ObjectDependencies, err = template.ParseTemplate(obj, container.Runtime.Configuration, format)

	if err != nil {
		return err
	}

	container.Runtime.ObjectDependencies = append(container.Runtime.ObjectDependencies, dependencyMap...)
	return nil
}

func (container *Container) PrepareResources(client *client.Http, user *authentication.User) error {
	for k, v := range container.Static.Resources.Resources {
		format := f.New("resource", v.Reference.Group, v.Reference.Name, "object")

		obj := objects.New(client.Get(user.Username), user)
		err := obj.Find(format)

		if err != nil {
			return errors.New(fmt.Sprintf("failed to fetch resource from the kv store %s", format.ToString()))
		}

		resourceObject := v1.ResourceDefinition{}

		err = json.Unmarshal(obj.GetDefinitionByte(), &resourceObject)

		if err != nil {
			return err
		}

		v.Docker.Data = resourceObject.Spec.Data
		container.Static.Resources.Resources[k].Docker.Data, _, err = template.ParseTemplate(obj, resourceObject.Spec.Data, nil)

		if err != nil {
			return err
		}

		tmpFile, err := os.CreateTemp("/tmp", container.Static.Name)
		if err != nil {
			log.Fatal(err)
		}

		if _, err = tmpFile.WriteString(UnpackSecretsResources(client, user, v.Docker.Data[v.Reference.Key])); err != nil {
			log.Fatal(err)
		}

		fmt.Println("RESOURCEEEEE")
		fmt.Println(v.Reference)
		fmt.Println(v.Reference.MountPoint)

		container.Static.Volumes.Add(v1.ContainerVolume{
			Type:       "bind",
			HostPath:   tmpFile.Name(),
			MountPoint: v.Reference.MountPoint,
		})

		container.Runtime.ObjectDependencies = append(container.Runtime.ObjectDependencies, &f.Format{
			Kind:       "resource",
			Group:      container.Static.Resources.Resources[k].Reference.Group,
			Identifier: container.Static.Resources.Resources[k].Reference.Name,
			Key:        "",
		})
	}

	return nil
}

func (container *Container) PrepareLabels() {
	for index, _ := range container.Static.Labels {
		regexDetectBigBrackets := regexp.MustCompile(`{{([^{\n}]*)}}`)
		matches := regexDetectBigBrackets.FindAllStringSubmatch(index, -1)

		if len(matches) > 0 {
			trimmedMatch := strings.TrimSpace(matches[0][1])
			SplitByDot := strings.SplitN(trimmedMatch, ".", 2)

			if len(SplitByDot) > 1 && container.Runtime.Configuration[SplitByDot[1]] != "" {
				newIndex := strings.Replace(index, matches[0][0], container.Runtime.Configuration[SplitByDot[1]], 1)
				container.Static.Labels[newIndex] = container.Static.Labels[index]

				delete(container.Static.Labels, index)
			}
		}
	}
}

func (container *Container) PrepareEnvs() {
	for index, value := range container.Static.Env {
		regexDetectBigBrackets := regexp.MustCompile(`{{([^{\n}]*)}}`)
		matches := regexDetectBigBrackets.FindAllStringSubmatch(value, -1)

		if len(matches) > 0 {
			SplitByDot := strings.SplitN(matches[0][1], ".", 2)

			trimmedIndex := strings.TrimSpace(SplitByDot[1])

			if len(SplitByDot) > 1 && container.Runtime.Configuration[trimmedIndex] != "" {
				container.Static.Env[index] = strings.Replace(container.Static.Env[index], matches[0][0], container.Runtime.Configuration[trimmedIndex], 1)
			}
		}
	}
}

func (container *Container) PrepareReadiness() {
	for indexReadiness, _ := range container.Static.Readiness.Readinesses {
		for index, value := range container.Static.Readiness.Readinesses[indexReadiness].Reference.Data {
			regexDetectBigBrackets := regexp.MustCompile(`{{([^{\n}]*)}}`)
			matches := regexDetectBigBrackets.FindAllStringSubmatch(value, -1)

			if len(matches) > 0 {
				format := f.NewFromString(matches[0][1])

				if format.IsValid() && format.Kind == "secret" {
					continue
				} else {
					container.Static.Readiness.Readinesses[indexReadiness].Docker.Body[index] = strings.Replace(container.Static.Readiness.Readinesses[indexReadiness].Docker.Body[index], matches[0][0], container.Runtime.Configuration[format.Group], 1)
				}
			}
		}
	}
}

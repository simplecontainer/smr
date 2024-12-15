package internal

import (
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
)

type Resources struct {
	Resources []*Resource
}

type Resource struct {
	Reference ResourceReference
	Docker    ResourceDocker
}

type ResourceReference struct {
	Group      string
	Name       string
	Key        string
	MountPoint string
}

type ResourceDocker struct {
	Data map[string]string
}

func NewResources(resources []v1.ContainerResource) *Resources {
	resourcesObj := &Resources{
		Resources: make([]*Resource, 0),
	}

	for _, resource := range resources {
		resourcesObj.Add(resource)
	}

	return resourcesObj
}

func NewResource(resource v1.ContainerResource) *Resource {
	return &Resource{
		Reference: ResourceReference{
			Group:      resource.Group,
			Name:       resource.Name,
			Key:        resource.Key,
			MountPoint: resource.MountPoint,
		},
		Docker: ResourceDocker{
			Data: make(map[string]string, 0),
		},
	}
}

func (resources *Resources) Add(resource v1.ContainerResource) {
	resources.Resources = append(resources.Resources, NewResource(resource))
}

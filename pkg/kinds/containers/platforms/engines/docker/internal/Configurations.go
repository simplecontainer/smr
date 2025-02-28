package internal

import (
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
)

type Configurations struct {
	Configurations []*Configuration
}

type Configuration struct {
	Reference ConfigurationReference
}

type ConfigurationReference struct {
	Group string
	Name  string
}

func NewConfigurations(configurationsDefinition []v1.ContainersConfigurations) *Configurations {
	configurationsObj := &Configurations{
		Configurations: make([]*Configuration, 0),
	}

	for _, configuration := range configurationsDefinition {
		configurationsObj.Add(configuration)
	}

	return configurationsObj
}

func NewConfiguration(configuration v1.ContainersConfigurations) *Configuration {
	return &Configuration{
		Reference: ConfigurationReference{
			Group: configuration.Group,
			Name:  configuration.Name,
		},
	}
}

func (configurations *Configurations) Add(Configuration v1.ContainersConfigurations) {
	configurations.Configurations = append(configurations.Configurations, NewConfiguration(Configuration))
}

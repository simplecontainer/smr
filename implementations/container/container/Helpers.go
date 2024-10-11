package container

import (
	"encoding/json"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/logger"
)

func convertResourcesDefinitionToResources(definition []map[string]string) []Resource {
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
		resources[i].Data = make(map[string]string)
	}

	return resources
}

func convertReadinessDefinitionToReadiness(readinessDefinition []v1.ContainerReadiness) []Readiness {
	var readiness = make([]Readiness, 0)

	for _, val := range readinessDefinition {
		readinessTmp := Readiness{
			Name:     val.Name,
			Operator: val.Operator,
			Timeout:  val.Timeout,
			Body:     val.Body,
			Solved:   false,
			Ctx:      nil,
			Cancel:   nil,
		}

		readiness = append(readiness, readinessTmp)
	}

	return readiness
}

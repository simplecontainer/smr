package formaters

import (
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/containers"
)

func ContainerBuilder(objects []json.RawMessage) ([]ContainerInformation, error) {
	var display = make([]ContainerInformation, 0)

	for _, obj := range objects {
		var container = make(map[string]interface{})
		err := json.Unmarshal(obj, &container)
		if err != nil {
			return nil, err
		}

		containerObj, err := containers.NewGhost(container)
		if err != nil {
			return nil, err
		}

		info := ContainerInformation{
			Group:         containerObj.GetGroup(),
			Name:          containerObj.GetName(),
			GeneratedName: containerObj.GetGeneratedName(),
			Image:         containerObj.GetImageWithTag(),
			ImageState:    containerObj.GetImageState().String(),
			IPs:           "",
			Ports:         "",
			Dependencies:  "",
			DockerState:   "",
			SmrState:      containerObj.GetStatus().State.State,
		}

		if containerObj.GetGlobalDefinition() != nil {
			for _, port := range containerObj.GetGlobalDefinition().Spec.Ports {
				if port.Host != "" {
					info.Ports += fmt.Sprintf("%s:%s, ", port.Host, port.Container)
				} else {
					info.Ports += fmt.Sprintf("%s, ", port.Container)
				}
			}

			if info.Ports == "" {
				info.Ports = "-"
			}

			for name, ip := range containerObj.GetNetwork() {
				info.IPs += fmt.Sprintf("%s (%s), ", ip.String(), name)
			}

			for _, u := range containerObj.GetGlobalDefinition().Spec.Dependencies {
				info.Dependencies += fmt.Sprintf("%s.%s ", u.Group, u.Name)
			}

			if info.Dependencies == "" {
				info.Dependencies = "-"
			}
		}

		if containerObj.GetEngineState() != "" {
			info.DockerState = fmt.Sprintf("%s (%s)", containerObj.GetEngineState(), container["Type"])
		} else {
			info.DockerState = "-"
		}

		info.LastUpdate = RoundAndFormatDuration(containerObj.GetStatus().LastUpdate)

		info.NodeURL = containerObj.GetRuntime().Node.URL
		info.NodeName = containerObj.GetRuntime().Node.NodeName
		info.NodeID = containerObj.GetRuntime().Node.NodeID

		display = append(display, info)
	}

	return display, nil
}

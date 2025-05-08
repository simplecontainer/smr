package formaters

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/containers"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/viper"
	"time"
)

type ContainerInformation struct {
	Group         string
	Name          string
	GeneratedName string
	Image         string
	Tag           string
	IPs           string
	Ports         string
	Dependencies  string
	DockerState   string
	SmrState      string
	NodeName      string
	NodeURL       string
	Recreated     bool
	LastUpdate    time.Duration
}

func Container(objects []json.RawMessage) {
	var display = make([]ContainerInformation, 0)

	for _, obj := range objects {
		var bytes []byte
		var err error

		var container = make(map[string]interface{})
		err = json.Unmarshal(obj, &container)

		switch container["Type"].(string) {
		case static.PLATFORM_DOCKER:
			ghost := &containers.Container{
				Platform: &docker.Docker{},
				General:  &containers.General{},
				Type:     static.PLATFORM_DOCKER,
			}

			bytes, err = json.Marshal(container)

			if err != nil {
				continue
			}

			err = json.Unmarshal(bytes, ghost)
			if err != nil {
				fmt.Println(err)
				continue
			}

			info := ContainerInformation{
				Group:         ghost.Platform.GetGroup(),
				Name:          ghost.Platform.GetName(),
				GeneratedName: ghost.Platform.GetGeneratedName(),
				Image:         ghost.Platform.(*docker.Docker).Image,
				Tag:           ghost.Platform.(*docker.Docker).Tag,
				IPs:           "",
				Ports:         "",
				Dependencies:  "",
				DockerState:   "",
				SmrState:      ghost.General.Status.State.State,
			}

			for _, port := range ghost.Platform.(*docker.Docker).Ports.Ports {
				if port.Host != "" {
					info.Ports += fmt.Sprintf("%s:%s, ", port.Host, port.Container)
				} else {
					info.Ports += fmt.Sprintf("%s, ", port.Container)
				}
			}

			if info.Ports == "" {
				info.Ports = "-"
			}

			for _, network := range ghost.Platform.(*docker.Docker).Networks.Networks {
				if network.Docker.IP != "" {
					info.IPs += fmt.Sprintf("%s (%s), ", network.Docker.IP, network.Reference.Name)
				}
			}

			for _, u := range ghost.Platform.(*docker.Docker).Definition.Spec.Dependencies {
				info.Dependencies += fmt.Sprintf("%s.%s ", u.Group, u.Name)
			}

			if info.Dependencies == "" {
				info.Dependencies = "-"
			}

			if ghost.Platform.(*docker.Docker).DockerState != "" {
				info.DockerState = fmt.Sprintf("%s (%s)", ghost.Platform.(*docker.Docker).DockerState, static.PLATFORM_DOCKER)
			} else {
				info.DockerState = "-"
			}

			info.LastUpdate = time.Since(ghost.GetStatus().LastUpdate).Round(time.Second)

			info.NodeURL = ghost.General.Runtime.Node.URL
			info.NodeName = ghost.General.Runtime.Node.NodeName

			display = append(display, info)
		}
	}

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	switch viper.GetString("o") {
	case "d":
		tbl := table.New("NODE", "RESOURCE", "IMAGE", "PORTS", "ENGINE STATE", "SMR STATE")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

		for _, container := range display {
			tbl.AddRow(
				fmt.Sprintf("%s", container.NodeName),
				fmt.Sprintf("%s/%s/%s", static.KIND_CONTAINERS, helpers.CliRemoveComa(container.Group), helpers.CliRemoveComa(container.GeneratedName)),
				fmt.Sprintf("%s:%s", container.Image, container.Tag),
				helpers.CliRemoveComa(container.Ports),
				container.DockerState,
				fmt.Sprintf("%s%s (%s)", container.SmrState, helpers.CliMask(container.Recreated, " (*)", ""), container.LastUpdate),
			)
		}

		tbl.Print()
		break
	case "s":
		tbl := table.New("NODE", "RESOURCE", "ENGINE STATE", "SMR STATE")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

		for _, container := range display {
			tbl.AddRow(
				fmt.Sprintf("%s", container.NodeName),
				fmt.Sprintf("%s/%s/%s", static.KIND_CONTAINERS, helpers.CliRemoveComa(container.Group), helpers.CliRemoveComa(container.GeneratedName)),
				container.DockerState,
				fmt.Sprintf("%s%s (%s)", container.SmrState, helpers.CliMask(container.Recreated, " (*)", ""), container.LastUpdate),
			)
		}

		tbl.Print()
		break
	}
}

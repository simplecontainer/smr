package formaters

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/containers"
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
		var container = make(map[string]interface{})
		err := json.Unmarshal(obj, &container)

		if err != nil {
			helpers.PrintAndExit(err, 1)
		}

		containerObj, err := containers.NewGhost(container)

		if err != nil {
			helpers.PrintAndExit(err, 1)
		}

		info := ContainerInformation{
			Group:         containerObj.GetGroup(),
			Name:          containerObj.GetName(),
			GeneratedName: containerObj.GetGeneratedName(),
			Image:         fmt.Sprintf("%s (%s)", containerObj.GetImageWithTag(), containerObj.GetImageState().String()),
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

		info.LastUpdate = time.Since(containerObj.GetStatus().LastUpdate).Round(time.Second)

		info.NodeURL = containerObj.GetRuntime().Node.URL
		info.NodeName = containerObj.GetRuntime().Node.NodeName

		display = append(display, info)
	}

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	switch viper.GetString("output") {
	case "full":
		tbl := table.New("NODE", "RESOURCE", "IMAGE", "PORTS", "ENGINE STATE", "SMR STATE")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

		for _, container := range display {
			tbl.AddRow(
				fmt.Sprintf("%s", container.NodeName),
				fmt.Sprintf("%s/%s/%s", static.KIND_CONTAINERS, helpers.CliRemoveComa(container.Group), helpers.CliRemoveComa(container.GeneratedName)),
				fmt.Sprintf("%s", container.Image),
				helpers.CliRemoveComa(container.Ports),
				container.DockerState,
				fmt.Sprintf("%s%s (%s)", container.SmrState, helpers.CliMask(container.Recreated, " (*)", ""), container.LastUpdate),
			)
		}

		tbl.Print()
		break
	case "short":
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

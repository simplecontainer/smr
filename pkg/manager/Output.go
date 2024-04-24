package manager

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/rodaine/table"
	"smr/pkg/cli"
	"sort"
)

func (mgr *Manager) OutputTable() {
	containers := cli.SendPs("http://localhost:8080/ps")

	keys := make([]string, 0, len(containers))

	for k := range containers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Name", "Image", "IPs", "Ports", "Dependencies", "Status")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, k := range keys {
		for _, v := range containers[k] {
			ips := ""
			ports := ""
			deps := ""
			status := ""

			for _, x := range v.Static.MappingPorts {
				ports += fmt.Sprintf("%s:%s ", x.Host, x.Container)
			}

			for _, u := range v.Runtime.Networks {
				ips += fmt.Sprintf("%s ", u.IP)
			}

			for _, u := range v.Static.Definition.Spec.Container.Dependencies {
				deps += fmt.Sprintf("%s ", u.Name)
			}

			if v.Status.DependsSolved {
				status += fmt.Sprintf("%s ", "Dependency solved")
			} else {
				status += fmt.Sprintf("%s ", "Dependency waiting")
			}

			tbl.AddRow(v.Static.Name, fmt.Sprintf("%s:%s", v.Static.Image, v.Static.Tag), ips, ports, deps, status)
		}
	}

	tbl.Print()
}

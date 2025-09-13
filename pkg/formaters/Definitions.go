package formaters

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/packer"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/viper"
	"os"
)

func Definitions(objects []json.RawMessage) {
	var gitopsObjs, err = GitopsBuilder(objects)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	switch viper.GetString("output") {
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"RESOURCE", "DEFINITIONS", "DRIFTED", "LAST SYNC"})

		SetStyle(table)

		for _, g := range gitopsObjs {
			for _, d := range g.Gitops.Pack.Definitions {
				table.Append([]string{
					fmt.Sprintf("%s/%s/%s", static.KIND_GITOPS, g.GetGroup(), g.GetName()),
					fmt.Sprintf("%s/%s/%s", d.Definition.Definition.GetKind(), d.Definition.Definition.GetMeta().GetGroup(), d.Definition.Definition.GetMeta().GetName()),
					helpers.CliMask(len(d.Definition.Definition.GetState().Gitops.Changes) > 0 || !d.Definition.Definition.GetState().Gitops.Synced, "Drifted", "InSync"),
					RoundAndFormatDuration(d.Definition.Definition.GetState().Gitops.LastSync),
				})
			}
		}

		table.Render()
		break
	case "json":
		var defs []*packer.Definition

		for _, g := range gitopsObjs {
			for _, d := range g.Gitops.Pack.Definitions {
				defs = append(defs, d)
			}
		}

		bytes, err := json.Marshal(defs)
		if err != nil {
			helpers.PrintAndExit(err, 1)
		}

		fmt.Println(string(bytes))
		break
	}
}

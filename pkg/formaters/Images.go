package formaters

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/static"
	"github.com/spf13/viper"
	"os"
)

func Images(objects []json.RawMessage) {
	display, err := ContainerBuilder(objects)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	switch viper.GetString("output") {
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"CONTAINER", "IMAGE", "IMAGE STATE"})

		SetStyle(table)

		for _, container := range display {
			table.Append([]string{
				fmt.Sprintf("%s/%s/%s", static.KIND_CONTAINERS, helpers.CliRemoveComa(container.Group), helpers.CliRemoveComa(container.GeneratedName)),
				container.Image,
				container.ImageState,
			})
		}

		table.Render()
		break
	case "json":
		bytes, err := json.MarshalIndent(objects, "", "  ")
		if err != nil {
			helpers.PrintAndExit(err, 1)
		}

		fmt.Println(string(bytes))
		break
	}
}

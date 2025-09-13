package formaters

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"os"
)

func Default(objects []json.RawMessage) {
	var definitions = make([]v1.CommonDefinition, 0)

	for _, obj := range objects {
		definition := v1.CommonDefinition{}

		err := json.Unmarshal(obj, &definition)
		if err != nil {
			continue
		}

		definitions = append(definitions, definition)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"RESOURCE"})

	SetStyle(table)

	for _, d := range definitions {
		table.Append([]string{fmt.Sprintf("%s/%s/%s", d.GetKind(), d.Meta.Group, d.Meta.Name)})
	}

	table.Render()
}

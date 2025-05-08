package formaters

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/rodaine/table"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
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

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("RESOURCE")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, d := range definitions {
		tbl.AddRow(fmt.Sprintf("%s/%s/%s", d.GetKind(), d.Meta.Group, d.Meta.Name))
	}

	tbl.Print()
}

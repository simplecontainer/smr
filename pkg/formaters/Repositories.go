package formaters

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/static"
	"os"
)

func Repositories(objects []json.RawMessage) {
	var gitopsObjs, err = GitopsBuilder(objects)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"RESOURCE", "REPOSITORY", "COMMIT", "REVISION"})

	SetStyle(table)

	for _, g := range gitopsObjs {
		table.Append([]string{
			fmt.Sprintf("%s/%s/%s", static.KIND_GITOPS, g.GetGroup(), g.GetName()),
			g.GetGit().Repository,
			helpers.CliMask(g.GetCommit() != nil && g.GetCommit().ID().IsZero(), "Not pulled", g.GetCommit().ID().String()[:7]),
			g.GetGit().Revision,
		})
	}

	table.Render()
}

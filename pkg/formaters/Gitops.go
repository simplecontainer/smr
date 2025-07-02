package formaters

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation"
	"github.com/simplecontainer/smr/pkg/static"
)

func Gitops(objects []json.RawMessage) {
	var gitopsObjs = make([]implementation.Gitops, 0)

	for _, obj := range objects {
		gitopsObj := implementation.Gitops{}

		err := json.Unmarshal(obj, &gitopsObj)

		if err != nil {
			fmt.Println(err)
			continue
		}

		gitopsObjs = append(gitopsObjs, gitopsObj)
	}

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("RESOURCE", "REPOSITORY", "REVISION", "SYNCED", "AUTO", "STATE", "STATUS")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for _, g := range gitopsObjs {
		certRef := fmt.Sprintf("%s.%s", g.Gitops.Auth.CertKeyRef.Group, g.Gitops.Auth.CertKeyRef.Name)
		httpRef := fmt.Sprintf("%s.%s", g.Gitops.Auth.HttpAuthRef.Group, g.Gitops.Auth.HttpAuthRef.Name)

		if certRef == "." {
			certRef = ""
		}

		if httpRef == "." {
			httpRef = ""
		}

		if g.Definition == nil {
			continue
		}

		if g.GetCommit() != nil {
			tbl.AddRow(
				fmt.Sprintf("%s/%s/%s", static.KIND_GITOPS, g.Definition.Meta.Group, g.Definition.Meta.Name),
				helpers.CliMask(g.GetCommit() != nil && g.GetCommit().ID().IsZero(), fmt.Sprintf("%s (Not pulled)", g.GetGit().Repository), fmt.Sprintf("%s (%s)", g.GetGit().Repository, g.GetCommit().ID().String()[:7])),
				g.GetGit().Revision,
				helpers.CliMask(g.GetStatus().LastSyncedCommit.IsZero(), "Never synced", g.GetStatus().LastSyncedCommit.String()[:7]),
				g.GetAutoSync(),
				helpers.CliMask(g.GetStatus().InSync, "InSync", "Drifted"),
				g.GetStatus().State.State,
			)
		} else {
			tbl.AddRow(
				fmt.Sprintf("%s/%s/%s", static.KIND_GITOPS, g.Definition.Meta.Group, g.Definition.Meta.Name),
				helpers.CliMask(g.GetCommit() != nil && g.GetCommit().ID().IsZero(), fmt.Sprintf("%s (Not pulled)", g.GetGit().Repository), fmt.Sprintf("%s", g.GetGit().Repository)),
				g.GetGit().Revision,
				helpers.CliMask(g.GetStatus().LastSyncedCommit.IsZero(), "Never synced", g.GetStatus().LastSyncedCommit.String()[:7]),
				g.GetAutoSync(),
				helpers.CliMask(g.GetStatus().InSync, "InSync", "Drifted"),
				g.GetStatus().State.State,
			)
		}
	}

	tbl.Print()
}

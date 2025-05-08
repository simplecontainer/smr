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
		certRef := fmt.Sprintf("%s.%s", g.Auth.CertKeyRef.Group, g.Auth.CertKeyRef.Name)
		httpRef := fmt.Sprintf("%s.%s", g.Auth.HttpAuthRef.Group, g.Auth.HttpAuthRef.Name)

		if certRef == "." {
			certRef = ""
		}

		if httpRef == "." {
			httpRef = ""
		}

		if g.Definition == nil {
			continue
		}

		if g.Commit != nil {
			tbl.AddRow(
				fmt.Sprintf("%s/%s/%s", static.KIND_GITOPS, g.Definition.Meta.Group, g.Definition.Meta.Name),
				helpers.CliMask(g.Commit != nil && g.Commit.ID().IsZero(), fmt.Sprintf("%s (Not pulled)", g.Git.Repository), fmt.Sprintf("%s (%s)", g.Git.Repository, g.Commit.ID().String()[:7])),
				g.Git.Revision,
				helpers.CliMask(g.Status.LastSyncedCommit.IsZero(), "Never synced", g.Status.LastSyncedCommit.String()[:7]),
				g.AutomaticSync,
				helpers.CliMask(g.Status.InSync, "InSync", "Drifted"),
				g.Status.State.State,
			)
		} else {
			tbl.AddRow(
				fmt.Sprintf("%s/%s/%s", static.KIND_GITOPS, g.Definition.Meta.Group, g.Definition.Meta.Name),
				helpers.CliMask(g.Commit != nil && g.Commit.ID().IsZero(), fmt.Sprintf("%s (Not pulled)", g.Git.Repository), fmt.Sprintf("%s", g.Git.Repository)),
				g.Git.Revision,
				helpers.CliMask(g.Status.LastSyncedCommit.IsZero(), "Never synced", g.Status.LastSyncedCommit.String()[:7]),
				g.AutomaticSync,
				helpers.CliMask(g.Status.InSync, "InSync", "Drifted"),
				g.Status.State.State,
			)
		}
	}

	tbl.Print()
}

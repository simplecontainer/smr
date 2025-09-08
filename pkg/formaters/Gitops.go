package formaters

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/static"
	"os"
)

func Gitops(objects []json.RawMessage) {
	var gitopsObjs, err = GitopsBuilder(objects)
	if err != nil {
		helpers.PrintAndExit(err, 1)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"RESOURCE", "REPOSITORY", "REVISION", "SYNCED", "AUTO SYNC", "STATUS"})

	SetStyle(table)

	for _, g := range gitopsObjs {
		var httpRef string
		var certRef string

		if g.Gitops.Auth.CertKeyRef != nil {
			certRef = fmt.Sprintf("%s.%s", g.Gitops.Auth.CertKeyRef.Group, g.Gitops.Auth.CertKeyRef.Name)
		}

		if g.Gitops.Auth.HttpAuthRef != nil {
			httpRef = fmt.Sprintf("%s.%s", g.Gitops.Auth.HttpAuthRef.Group, g.Gitops.Auth.HttpAuthRef.Name)
		}

		if certRef == "." {
			certRef = ""
		}

		if httpRef == "." {
			httpRef = ""
		}

		if g.GetCommit() != nil {
			table.Append([]string{
				fmt.Sprintf("%s/%s/%s", static.KIND_GITOPS, g.GetGroup(), g.GetName()),
				helpers.CliMask(g.GetCommit() != nil && g.GetCommit().ID().IsZero(), fmt.Sprintf("%s (Not pulled)", g.GetGit().Repository), fmt.Sprintf("%s (%s)", g.GetGit().Repository, g.GetCommit().ID().String()[:7])),
				g.GetGit().Revision,
				helpers.CliMask(g.GetStatus().LastSyncedCommit.IsZero(), "Never synced", g.GetStatus().LastSyncedCommit.String()[:7]),
				fmt.Sprintf("%v", g.GetAutoSync()),
				g.GetStatus().State.State,
			})
		} else {
			table.Append([]string{
				fmt.Sprintf("%s/%s/%s", static.KIND_GITOPS, g.GetGroup(), g.GetName()),
				helpers.CliMask(g.GetCommit() != nil && g.GetCommit().ID().IsZero(), fmt.Sprintf("%s (Not pulled)", g.GetGit().Repository), fmt.Sprintf("%s", g.GetGit().Repository)),
				g.GetGit().Revision,
				helpers.CliMask(g.GetStatus().LastSyncedCommit.IsZero(), "Never synced", g.GetStatus().LastSyncedCommit.String()[:7]),
				fmt.Sprintf("%v", g.GetAutoSync()),
				g.GetStatus().State.State,
			})
		}
	}

	table.Render()
}

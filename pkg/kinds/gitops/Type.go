package gitops

import (
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
	"github.com/simplecontainer/smr/pkg/static"
)

type Gitops struct {
	Started bool
	Shared  *shared.Shared
}

const KIND string = static.KIND_GITOPS

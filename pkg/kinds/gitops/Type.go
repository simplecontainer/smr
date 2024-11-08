package gitops

import (
	"github.com/simplecontainer/smr/pkg/kinds/gitops/shared"
)

type Gitops struct {
	Started bool
	Shared  *shared.Shared
}

const KIND string = "gitops"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}

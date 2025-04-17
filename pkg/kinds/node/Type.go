package node

import (
	"github.com/simplecontainer/smr/pkg/kinds/node/shared"
	"github.com/simplecontainer/smr/pkg/static"
)

type Node struct {
	Started bool
	Shared  *shared.Shared
}

const KIND string = static.KIND_CERTKEY

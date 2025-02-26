package ishared

import "github.com/simplecontainer/smr/pkg/manager"

type Shared interface {
	GetManager() *manager.Manager
}

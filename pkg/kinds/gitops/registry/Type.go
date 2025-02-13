package registry

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation"
	"sync"
)

type Registry struct {
	Gitops     map[string]*implementation.Gitops
	GitopsLock sync.RWMutex
	Client     *client.Http
	User       *authentication.User
}

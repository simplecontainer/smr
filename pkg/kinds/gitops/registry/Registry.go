package registry

import (
	"errors"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
)

func New(client *clients.Http, user *authentication.User) *Registry {
	return &Registry{
		Gitops: make(map[string]*implementation.Gitops),
		Client: client,
		User:   user,
	}
}

func (registry *Registry) AddOrUpdate(group string, name string, gitops *implementation.Gitops) {
	registry.GitopsLock.Lock()
	registry.Gitops[common.GroupIdentifier(group, name)] = gitops
	registry.GitopsLock.Unlock()
}

func (registry *Registry) Remove(prefix string, group string, name string) error {
	registry.GitopsLock.Lock()
	defer registry.GitopsLock.Unlock()

	if registry.Gitops[common.GroupIdentifier(group, name)] == nil {
		return errors.New("gitops not found")
	} else {
		delete(registry.Gitops, common.GroupIdentifier(group, name))

		format := f.New(prefix, static.CATEGORY_STATE, static.KIND_GITOPS, group, name, name)
		obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

		err := obj.Propose(format, nil)

		if err != nil {
			return err
		}

		return nil
	}
}

func (registry *Registry) FindLocal(group string, name string) *implementation.Gitops {
	registry.GitopsLock.RLock()
	defer registry.GitopsLock.RUnlock()

	value, ok := registry.Gitops[common.GroupIdentifier(group, name)]

	if ok {
		return value
	} else {
		return nil
	}
}

func (registry *Registry) Sync(group string, name string) error {
	registry.GitopsLock.RLock()
	gitopsObj, ok := registry.Gitops[common.GroupIdentifier(group, name)]
	registry.GitopsLock.RUnlock()

	if ok {
		format := f.New(gitopsObj.GetDefinition().GetPrefix(), static.CATEGORY_STATE, static.KIND_GITOPS, gitopsObj.GetGroup(), gitopsObj.GetName(), gitopsObj.GetName())
		obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

		bytes, err := gitopsObj.ToJSON()

		if err != nil {
			return err
		}

		return obj.Wait(format, bytes)
	} else {
		return errors.New("gitops not found on this node")
	}
}

package registry

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"time"
)

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

		format := f.New(prefix, static.CATEGORY_STATE, static.KIND_GITOPS, group, name)
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

	fmt.Println("syncing registry", group, name, time.Now())

	if ok {
		format := f.New(gitopsObj.Definition.GetPrefix(), static.CATEGORY_STATE, static.KIND_GITOPS, gitopsObj.GetGroup(), gitopsObj.GetName())
		obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

		bytes, err := gitopsObj.ToJson()

		if err != nil {
			return err
		}

		return obj.Wait(format, bytes)
	} else {
		return errors.New("gitops not found on this node")
	}
}

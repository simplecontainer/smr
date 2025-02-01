package registry

import (
	"errors"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
)

func (registry *Registry) AddOrUpdate(group string, name string, gitops *implementation.Gitops) {
	registry.GitopsLock.Lock()

	if registry.Gitopses[group] == nil {
		tmp := make(map[string]*implementation.Gitops)
		tmp[name] = gitops

		registry.Gitopses[group] = tmp
	} else {
		registry.Gitopses[group][name] = gitops
	}

	registry.GitopsLock.Unlock()
}

func (registry *Registry) Sync(gitops *implementation.Gitops) error {
	format := f.New(gitops.Definition.GetPrefix(), static.CATEGORY_STATE, static.KIND_GITOPS, gitops.Definition.Meta.Group, gitops.Definition.Meta.Name)
	obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

	bytes, err := gitops.ToJson()

	if err != nil {
		return err
	}

	return obj.Wait(format, bytes)
}

func (registry *Registry) Remove(prefix string, group string, name string) error {
	registry.GitopsLock.Lock()
	defer registry.GitopsLock.Unlock()

	if registry.Gitopses[group] == nil {
		return errors.New("gitops not found")
	} else {
		delete(registry.Gitopses[group], name)

		if len(registry.Gitopses[group]) == 0 {
			delete(registry.Gitopses, group)
		}

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

	if registry.Gitopses[group] != nil {
		if registry.Gitopses[group][name] != nil {
			return registry.Gitopses[group][name]
		} else {
			return nil
		}
	} else {
		return nil
	}
}

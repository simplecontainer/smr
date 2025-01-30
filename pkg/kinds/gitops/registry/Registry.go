package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/gitops/implementation"
	"github.com/simplecontainer/smr/pkg/logger"
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
	format := f.NewUnformated(fmt.Sprintf("state.gitops.%s.%s", gitops.Definition.Meta.Group, gitops.Definition.Meta.Name), static.CATEGORY_PLAIN_STRING)
	obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

	bytes, err := gitops.ToJson()

	if err != nil {
		return err
	}

	return obj.Wait(format, bytes)
}

func (registry *Registry) Remove(group string, name string) error {
	registry.GitopsLock.Lock()
	defer registry.GitopsLock.Unlock()

	if registry.Gitopses[group] == nil {
		return errors.New("gitops not found")
	} else {
		delete(registry.Gitopses[group], name)

		if len(registry.Gitopses[group]) == 0 {
			delete(registry.Gitopses, group)
		}

		format := f.NewUnformated(fmt.Sprintf("state.gitops.%s.%s", group, name), static.CATEGORY_PLAIN_STRING)
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

func (registry *Registry) Find(group string, name string) *implementation.Gitops {
	format := f.NewUnformated(fmt.Sprintf("state.gitops.%s.%s", group, name), static.CATEGORY_PLAIN_STRING)
	obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

	registry.GitopsLock.RLock()

	if registry.Gitopses[group] != nil && registry.Gitopses[group][name] != nil {
		registry.GitopsLock.RUnlock()
		return registry.Gitopses[group][name]
	} else {
		registry.GitopsLock.RUnlock()

		obj.Find(format)

		if obj.Exists() {
			instance := &implementation.Gitops{}
			err := json.Unmarshal(obj.GetDefinitionByte(), instance)

			if err != nil {
				return nil
			}

			instance.Ghost = true

			if err != nil {
				return nil
			}

			return instance
		}

		return nil
	}
}

func (registry *Registry) All() map[string]map[string]*implementation.Gitops {
	format := f.NewUnformated("state.gitops", static.CATEGORY_PLAIN_STRING)
	obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

	var result = make(map[string]map[string]*implementation.Gitops)
	objs, _ := obj.FindMany(format)

	if len(objs) > 0 {
		for _, o := range objs {
			instance := &implementation.Gitops{}
			err := json.Unmarshal(o.GetDefinitionByte(), instance)

			if err != nil {
				logger.Log.Error(err.Error())
				continue
			}

			if result[instance.Definition.Meta.Group] != nil {
				result[instance.Definition.Meta.Group][instance.Definition.Meta.Name] = instance
			} else {
				tmp := make(map[string]*implementation.Gitops)
				tmp[instance.Definition.Meta.Name] = instance

				result[instance.Definition.Meta.Group] = tmp
			}
		}
	}

	return result
}

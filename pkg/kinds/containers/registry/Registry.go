package registry

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/containers"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"strings"
	"time"
)

func New(client *clients.Http, user *authentication.User) platforms.Registry {
	return &Registry{
		Containers:     make(map[string]platforms.IContainer),
		BackOffTracker: make(map[string]*RestartTracker),
		Client:         client,
		User:           user,
	}
}

func (registry *Registry) AddOrUpdate(group string, name string, containerAddr platforms.IContainer) {
	registry.ContainersLock.Lock()
	defer registry.ContainersLock.Unlock()

	registry.Containers[common.GroupIdentifier(group, name)] = containerAddr
}

func (registry *Registry) Remove(prefix string, group string, name string) error {
	registry.ContainersLock.Lock()
	identifier := common.GroupIdentifier(group, name)

	if registry.Containers[identifier] == nil {
		registry.ContainersLock.Unlock()
		return errors.New(fmt.Sprintf("container not found: %s", identifier))
	}

	format := f.New(prefix, static.CATEGORY_STATE, static.KIND_CONTAINERS, group, registry.extractName(name), name)
	registry.ContainersLock.Unlock()

	obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)
	err := obj.Wait(format, nil)

	if err != nil {
		return err
	}

	registry.ContainersLock.Lock()
	delete(registry.Containers, identifier)
	delete(registry.BackOffTracker, identifier)
	registry.ContainersLock.Unlock()

	return nil
}

func (registry *Registry) FindLocal(group string, name string) platforms.IContainer {
	registry.ContainersLock.RLock()
	defer registry.ContainersLock.RUnlock()

	value, ok := registry.Containers[common.GroupIdentifier(group, name)]

	if ok {
		return value
	} else {
		return nil
	}
}

func (registry *Registry) Find(prefix string, group string, name string) platforms.IContainer {
	format := f.New(prefix, static.CATEGORY_STATE, static.KIND_CONTAINERS, group, registry.extractName(name), name)
	obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

	registry.ContainersLock.RLock()
	value, ok := registry.Containers[common.GroupIdentifier(group, name)]
	registry.ContainersLock.RUnlock()

	if ok {
		return value
	} else {
		obj.Find(format)

		if obj.Exists() {
			instance, err := containers.NewGhost(obj.GetDefinition())

			if err != nil {
				logger.Log.Error(err.Error())
				return nil
			}

			return instance
		} else {
			return nil
		}
	}
}

func (registry *Registry) FindGroup(prefix string, group string) []platforms.IContainer {
	format := f.New(prefix, static.CATEGORY_STATE, static.KIND_CONTAINERS, group)
	obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

	var result []platforms.IContainer
	objs, _ := obj.FindMany(format)

	if len(objs) > 0 {
		for _, o := range objs {
			instance, err := containers.NewGhost(o.GetDefinition())

			if err != nil {
				logger.Log.Error(err.Error())
				continue
			}

			result = append(result, instance)
		}
	}

	return result
}

func (registry *Registry) Sync(group string, name string) error {
	registry.ContainersLock.RLock()
	container, ok := registry.Containers[common.GroupIdentifier(group, name)]
	registry.ContainersLock.RUnlock()

	if ok {
		format := f.New(container.GetDefinition().GetPrefix(), static.CATEGORY_STATE, static.KIND_CONTAINERS, container.GetGroup(), container.GetName(), container.GetGeneratedName())
		obj := objects.New(registry.Client.Clients[registry.User.Username], registry.User)

		bytes, err := container.ToJSON()

		if err != nil {
			return err
		}

		return obj.Wait(format, bytes)
	} else {
		return errors.New("container not found on this node")
	}
}

func (registry *Registry) MarkContainerStarted(group string, name string) {
	registry.ContainersLock.Lock()
	defer registry.ContainersLock.Unlock()

	identifier := common.GroupIdentifier(group, name)

	if registry.BackOffTracker[identifier] == nil {
		registry.BackOffTracker[identifier] = &RestartTracker{
			RestartTimes: make([]time.Time, 0),
		}
	}

	registry.BackOffTracker[identifier].LastStartTime = time.Now()
	registry.BackOffTracker[identifier].IsRunning = true
}

func (registry *Registry) MarkContainerStopped(group string, name string) error {
	registry.ContainersLock.Lock()
	defer registry.ContainersLock.Unlock()

	identifier := common.GroupIdentifier(group, name)
	now := time.Now()

	if registry.BackOffTracker[identifier] == nil {
		registry.BackOffTracker[identifier] = &RestartTracker{
			RestartTimes: make([]time.Time, 0),
		}
	}

	tracker := registry.BackOffTracker[identifier]
	tracker.IsRunning = false

	runtime := now.Sub(tracker.LastStartTime)

	if runtime < MinHealthyRuntime {
		tracker.Count += 1

		if tracker.Count > MaxBackoffAttempts {
			return fmt.Errorf("container %s exceeded max immediate restart attempts (%d) - container keeps crashing",
				identifier, MaxBackoffAttempts)
		}
	} else {
		tracker.Count = 0
	}

	tracker.RestartTimes = append(tracker.RestartTimes, now)

	cutoff := now.Add(-RestartWindow)
	validRestarts := make([]time.Time, 0)
	for _, rt := range tracker.RestartTimes {
		if rt.After(cutoff) {
			validRestarts = append(validRestarts, rt)
		}
	}
	tracker.RestartTimes = validRestarts

	if len(tracker.RestartTimes) > MaxRestartsInWindow {
		return fmt.Errorf("container %s exceeded max restarts (%d) in %v - crash loop detected",
			identifier, MaxRestartsInWindow, RestartWindow)
	}

	return nil
}

func (registry *Registry) BackOff(group string, name string) error {
	registry.ContainersLock.Lock()
	defer registry.ContainersLock.Unlock()

	identifier := common.GroupIdentifier(group, name)

	if registry.BackOffTracker[identifier] == nil {
		registry.BackOffTracker[identifier] = &RestartTracker{
			RestartTimes: make([]time.Time, 0),
		}
	}

	registry.BackOffTracker[identifier].Count += 1

	if registry.BackOffTracker[identifier].Count > MaxBackoffAttempts {
		return fmt.Errorf("container %s exceeded max restart attempts (%d)", identifier, MaxBackoffAttempts)
	}

	return nil
}

func (registry *Registry) BackOffReset(group string, name string) {
	registry.ContainersLock.Lock()
	defer registry.ContainersLock.Unlock()

	identifier := common.GroupIdentifier(group, name)

	if registry.BackOffTracker[identifier] == nil {
		registry.BackOffTracker[identifier] = &RestartTracker{
			RestartTimes: make([]time.Time, 0),
		}
	}

	registry.BackOffTracker[identifier].Count = 0
}

func (registry *Registry) ResetRestartTracking(group string, name string) {
	registry.ContainersLock.Lock()
	defer registry.ContainersLock.Unlock()

	identifier := common.GroupIdentifier(group, name)
	registry.BackOffTracker[identifier] = &RestartTracker{
		RestartTimes: make([]time.Time, 0),
	}
}

func (registry *Registry) GetBackOffCount(group string, name string) uint64 {
	registry.ContainersLock.RLock()
	defer registry.ContainersLock.RUnlock()

	identifier := common.GroupIdentifier(group, name)
	if registry.BackOffTracker[identifier] == nil {
		return 0
	}
	return registry.BackOffTracker[identifier].Count
}

func (registry *Registry) GetRestartCount(group string, name string) int {
	registry.ContainersLock.RLock()
	defer registry.ContainersLock.RUnlock()

	identifier := common.GroupIdentifier(group, name)
	if registry.BackOffTracker[identifier] == nil {
		return 0
	}

	now := time.Now()
	cutoff := now.Add(-RestartWindow)
	count := 0

	for _, rt := range registry.BackOffTracker[identifier].RestartTimes {
		if rt.After(cutoff) {
			count++
		}
	}

	return count
}

func (registry *Registry) extractName(generatedName string) string {
	tmp := strings.Split(generatedName, "-")
	return strings.Join(tmp[1:len(tmp)-1], "-")
}

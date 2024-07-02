package manager

import "github.com/qdnqn/smr/pkg/reconciler"

func (mgr *Manager) EmitChange(kind string, group string, identifier string) {
	if mgr.Registry.Containers[group] != nil {
		if identifier == "*" {
			for identifierFromRegistry, container := range mgr.Registry.Containers[group] {
				if container.HasDependencyOn(kind, group, identifier) {
					mgr.Reconciler.QueueEvents <- reconciler.Events{
						Container: mgr.Registry.Containers[group][identifierFromRegistry],
						Kind:      "change",
						Message:   "detected change in dependent resource",
					}
				}
			}
		} else {
			if mgr.Registry.Containers[group][identifier] != nil {
				if mgr.Registry.Containers[group][identifier].HasDependencyOn(kind, group, identifier) {
					mgr.Reconciler.QueueEvents <- reconciler.Events{
						Container: mgr.Registry.Containers[group][identifier],
						Kind:      "change",
						Message:   "detected change in dependent resource",
					}
				}
			}
		}
	}
}

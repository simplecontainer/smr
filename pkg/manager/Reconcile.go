package manager

import "smr/pkg/reconciler"

func (mgr *Manager) Reconcile() {
	go mgr.Reconciler.ListenQueue(mgr.Registry, mgr.Runtime, mgr.Badger, mgr.DnsCache)
	go mgr.Reconciler.ListenDockerEvents(mgr.Registry, mgr.DnsCache)
	go mgr.Reconciler.ListenEvents(mgr.Registry, mgr.DnsCache)
}

func (mgr *Manager) EmitChange(group string, identifier string) {
	if mgr.Registry.Containers[group] != nil {
		if identifier == "*" {
			for identifierFromRegistry, _ := range mgr.Registry.Containers[group] {
				mgr.Reconciler.QueueEvents <- reconciler.Events{
					Container: mgr.Registry.Containers[group][identifierFromRegistry],
					Kind:      "change",
					Message:   "detected change in dependent resource",
				}
			}
		} else {
			if mgr.Registry.Containers[group][identifier] != nil {
				mgr.Reconciler.QueueEvents <- reconciler.Events{
					Container: mgr.Registry.Containers[group][identifier],
					Kind:      "change",
					Message:   "detected change in dependent resource",
				}
			}
		}
	}
}

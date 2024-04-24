package manager

func (mgr *Manager) Reconcile() {
	go mgr.Reconciler.ListenQueue(mgr.Registry, mgr.Runtime, mgr.Badger, mgr.DnsCache)
	go mgr.Reconciler.ListenEvents(mgr.Registry, mgr.DnsCache)
}

func (mgr *Manager) ConfigChangeEmit(group string, identifier string) {
	/*if mgr.Registry.Containers[group] != nil {
		if identifier == "*" {
			for identifierFromRegistry, _ := range mgr.Registry.Containers[group] {
				mgr.Registry.Containers[group][identifierFromRegistry].Status.Reconciling = true
				mgr.HandleReconcile(mgr.Registry.Containers[group][identifierFromRegistry])
			}
		}	else {
			if mgr.Registry.Containers[group][identifier] != nil {
				mgr.Registry.Containers[group][identifier].Status.Reconciling = true
				mgr.HandleReconcile(mgr.Registry.Containers[group][identifier])
			}
		}
	}

	*/
}

func (mgr *Manager) ResourceChangeEmit(group string, identifier string) {
	/*
		if mgr.Registry.Containers[group] != nil {
			if identifier == "*" {
				for identifierFromRegistry, _ := range mgr.Registry.Containers[group] {
					if mgr.Registry.Containers[group][identifierFromRegistry].Status.Running {
						mgr.Registry.Containers[group][identifierFromRegistry].Status.Reconciling = true
						mgr.HandleReconcile(mgr.Registry.Containers[group][identifierFromRegistry])
					}
				}
			}	else {
				if mgr.Registry.Containers[group][identifier] != nil {
					if mgr.Registry.Containers[group][identifier].Status.Running {
						mgr.Registry.Containers[group][identifier].Status.Reconciling = true
						mgr.HandleReconcile(mgr.Registry.Containers[group][identifier])
					}
				}
			}
		}
	*/
}

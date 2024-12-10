package registry

func (registry *Registry) ListenChanges() {
	for {
		select {
		case container, ok := <-registry.ChangeC:
			if ok {
				go registry.Sync(container)
			}
			break
		}
	}
}

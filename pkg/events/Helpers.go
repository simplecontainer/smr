package events

func reconcileIgnore(labels map[string]string) bool {
	val, exists := labels["reconcile"]

	if exists {
		if val == "false" {
			return true
		}
	}

	return false
}

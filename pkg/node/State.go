package node

func NewState() State {
	return State{
		Health: Health{
			Cluster:        false,
			Etcd:           false,
			Running:        false,
			MemoryPressure: false,
			CPUPressure:    false,
		},
		Control: Control{
			Upgrading:  false,
			Draining:   false,
			Recovering: false,
		},
	}
}

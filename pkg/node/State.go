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
			Starting:   StatusNotStarted,
			Upgrading:  StatusNotStarted,
			Draining:   StatusNotStarted,
			Recovering: StatusNotStarted,
		},
	}
}

func (s *State) ModifyControl(field string, status ControlStatus) {
	switch field {
	case "upgrading":
		s.Control.Upgrading = status
	case "draining":
		s.Control.Draining = status
	case "recovering":
		s.Control.Recovering = status
	}
}

func (s *State) ResetControl() {
	s.Control = Control{
		Starting:   StatusNotStarted,
		Upgrading:  StatusNotStarted,
		Draining:   StatusNotStarted,
		Recovering: StatusNotStarted,
	}
}

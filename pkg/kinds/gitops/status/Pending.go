package status

import (
	"errors"
)

func NewPending() *Pending {
	return &Pending{}
}

func (p *Pending) Set(state string) error {
	switch state {
	case PENDING_DELETE:
		p.Pending = state
		return nil
	case PENDING_SYNC:
		p.Pending = state
		return nil
	default:
		return errors.New("invalid pending state")
	}
}

func (p *Pending) Is(states ...string) bool {
	for _, state := range states {
		if state == p.Pending {
			return true
		}
	}

	return false
}

func (p *Pending) Clear() {
	p.Pending = ""
}

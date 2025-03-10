package status

import (
	"errors"
)

func NewPending() *Pending {
	return &Pending{}
}

func (p *Pending) Set(state string) error {
	if p.Pending != "" {
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
	} else {
		return errors.New("pending state is not empty - clear old one first")
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

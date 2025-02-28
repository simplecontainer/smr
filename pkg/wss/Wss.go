package wss

import (
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"sync"
)

func New() *WebSockets {
	return &WebSockets{
		Channels: make(map[int]chan ievents.Event, 0),
		Lock:     &sync.RWMutex{},
	}
}

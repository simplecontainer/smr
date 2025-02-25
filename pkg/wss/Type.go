package wss

import (
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"sync"
)

type WebSockets struct {
	Channels []chan ievents.Event
	Lock     *sync.RWMutex
}

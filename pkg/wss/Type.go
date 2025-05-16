package wss

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"sync"
)

type WebSockets struct {
	Channels map[int]chan ievents.Event
	Lock     *sync.RWMutex

	Connections map[string]*ConnectionControl
}

type ConnectionControl struct {
	Conn     *websocket.Conn
	ctx      context.Context
	Cancel   context.CancelFunc
	Mutex    *sync.Mutex
	IsClosed bool
}

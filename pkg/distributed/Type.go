package distributed

import (
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/smaps"
	"sync"
)

type Replication struct {
	Client      *clients.Client
	User        *authentication.User
	Node        *node.Node
	DataC       chan KV.KV
	EventsC     chan KV.KV
	Informer    *Informer
	DnsUpdatesC chan KV.KV
	Replicated  *smaps.Smap
}

type Informer struct {
	Chs  map[string]map[string]chan ievents.Event
	Lock *sync.RWMutex
}

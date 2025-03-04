package distributed

import (
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/smaps"
	"sync"
)

type Replication struct {
	Client      *client.Client
	User        *authentication.User
	NodeName    string
	Node        uint64
	DataC       chan KV.KV
	EventsC     chan KV.KV
	Informer    *Informer
	DnsUpdatesC chan KV.KV
	Replicated  *smaps.Smap
}

type Informer struct {
	Chs  map[string]chan ievents.Event
	Lock *sync.RWMutex
}

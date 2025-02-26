package distributed

import (
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts/ievents"
	"github.com/simplecontainer/smr/pkg/smaps"
)

type Replication struct {
	Client      *client.Client
	User        *authentication.User
	NodeName    string
	Node        uint64
	DataC       chan KV.KV
	EventsC     chan KV.KV
	DeleteC     map[string]chan ievents.Event
	DnsUpdatesC chan KV.KV
	Replicated  *smaps.Smap
}

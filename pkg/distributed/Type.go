package distributed

import (
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/smaps"
)

type Replication struct {
	Client      *client.Client
	User        *authentication.User
	NodeName    string
	DataC       chan KV.KV
	EventsC     chan KV.KV
	DnsUpdatesC chan KV.KV
	Replicated  *smaps.Smap
}

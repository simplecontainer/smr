package distributed

import (
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
)

type Replication struct {
	Client      *client.Client
	User        *authentication.User
	Node        string
	DataC       chan KV.KV
	EventsC     chan KV.KV
	DnsUpdatesC chan KV.KV
}

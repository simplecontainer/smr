package distributed

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
)

type Replication struct {
	Client      *client.Client
	User        *authentication.User
	Node        string
	DataC       chan KV
	EventsC     chan KV
	DnsUpdatesC chan KV
}

type KV struct {
	Key      string
	Val      []byte
	Category int
	Node     uint64
	Local    bool
}

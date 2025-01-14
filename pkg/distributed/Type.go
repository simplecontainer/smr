package distributed

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
)

type Replication struct {
	Client  *client.Client
	User    *authentication.User
	DataC   chan KV
	EventsC chan KV
}

type KV struct {
	Key      string
	Val      []byte
	Category int
	Agent    string
	Local    bool
}

package api

import (
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/distributed"
)

func (a *Api) SetupReplication() {
	a.Replication = distributed.New(a.Manager.Http.Clients[a.User.Username], a.User, a.Cluster.Node)
	a.Replication.EventsC = make(chan KV.KV)
	a.Replication.DnsUpdatesC = a.DnsCache.Records

	a.Manager.Replication = a.Replication
}

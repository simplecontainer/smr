package dns

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/distributed"
)

type Records struct {
	ARecords map[string]*ARecord
	Agent    string
	Client   *client.Http
	User     *authentication.User
	Updates  chan distributed.KV
}

type ARecord struct {
	IPs []string
}

type Distributed struct {
	Domain string
	IP     string
	Action uint8
}

const ADD_RECORD = 0x1
const REMOVE_RECORD = 0x2

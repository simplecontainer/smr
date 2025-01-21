package dns

import (
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/smaps"
)

type Records struct {
	ARecords    *smaps.Smap
	Client      *client.Http
	User        *authentication.User
	Nameservers []string
	Records     chan KV.KV
}

type ARecord struct {
	Addresses []string
}

type Distributed struct {
	Domain   string
	Headless string
	IP       string
	Action   uint8
}

const AddRecord = 0x1
const RemoveRecord = 0x2

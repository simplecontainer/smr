package dns

import (
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/smaps"
	"sync"
)

type Records struct {
	ARecords    *smaps.Smap
	Client      *clients.Http
	User        *authentication.User
	Lock        *sync.RWMutex
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

package dns

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
)

type Records struct {
	ARecords map[string]*ARecord
	Agent    string
	Client   *client.Http
	User     *authentication.User
}

type ARecord struct {
	IPs []string
}

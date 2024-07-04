package manager

import (
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/objectdependency"
)

type Manager struct {
	Config             *configuration.Configuration
	Keys               *keys.Keys
	DefinitionRegistry *objectdependency.DefinitionRegistry
	DnsCache           *dns.Records
}

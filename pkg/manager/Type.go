package manager

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/distributed"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/relations"
	"go.uber.org/zap/zapcore"
)

type Manager struct {
	User            *authentication.User
	Config          *configuration.Configuration
	Cluster         *cluster.Cluster
	Replication     *distributed.Replication
	Keys            *keys.Keys
	Kinds           *relations.RelationRegistry
	KindsRegistry   map[string]contracts.Kind
	PluginsRegistry []string
	Http            *client.Http
	DnsCache        *dns.Records
	LogLevel        zapcore.Level
	VersionServer   string
}

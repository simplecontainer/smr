package manager

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/relations"
	"go.uber.org/zap/zapcore"
)

type Manager struct {
	User             *authentication.User
	Config           *configuration.Configuration
	Keys             *keys.Keys
	RelationRegistry *relations.RelationRegistry
	PluginsRegistry  []string
	Http             *client.Http
	DnsCache         *dns.Records
	LogLevel         zapcore.Level
}

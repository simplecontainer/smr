package manager

import (
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/relations"
	"go.uber.org/zap/zapcore"
)

type Manager struct {
	Config           *configuration.Configuration
	Keys             *keys.Keys
	RelationRegistry *relations.RelationRegistry
	DnsCache         *dns.Records
	LogLevel         zapcore.Level
}

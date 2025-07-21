package client

import (
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contexts"
	"github.com/simplecontainer/smr/pkg/version"
)

type Client struct {
	Config  *configuration.Configuration
	Group   string
	Manager *contexts.Manager
	Context *contexts.ClientContext
	Version *version.VersionClient
}

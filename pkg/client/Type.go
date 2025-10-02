package client

import (
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contexts"
	"github.com/simplecontainer/smr/pkg/packer/ocicredentials"
	"github.com/simplecontainer/smr/pkg/packer/signature"
	"github.com/simplecontainer/smr/pkg/version"
)

type Client struct {
	Config      *configuration.Configuration
	Group       string
	Manager     *contexts.Manager
	Credentials *ocicredentials.Manager
	Signer      *signature.Signer
	Context     *contexts.ClientContext
	Version     *version.VersionClient
}

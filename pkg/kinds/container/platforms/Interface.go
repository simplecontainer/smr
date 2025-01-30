package platforms

import (
	TDTypes "github.com/docker/docker/api/types"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/container/status"
	"io"
)

type IContainer interface {
	Start() error
	Stop(signal string) error
	Kill(signal string) error
	Restart() error
	Delete() error
	Rename(newName string) error
	Exec(command []string) (types.ExecResult, error)
	Logs(bool) (io.ReadCloser, error)

	GetContainerState() (string, error)
	Run() (*TDTypes.Container, error)
	Prepare(config *configuration.Configuration, client *client.Http, user *authentication.User) error
	PostRun(config *configuration.Configuration, dnsCache *dns.Records) error

	AttachToNetworks(string) error
	UpdateDns(dnsCache *dns.Records) error
	RemoveDns(dnsCache *dns.Records, networkId string) error
	SyncNetworkInformation() error

	HasDependencyOn(string, string, string) bool
	HasOwner() bool

	GetRuntime() *types.Runtime
	GetStatus() *status.Status
	GetNode() uint64
	GetNodeName() string

	GetId() string
	GetDefinition() contracts.IDefinition
	GetLabels() map[string]string
	GetGeneratedName() string
	GetName() string
	GetGroup() string
	GetGroupIdentifier() string

	GetDomain(network string) string
	GetHeadlessDomain(network string) string

	IsGhost() bool
	SetGhost(bool)

	ToJson() ([]byte, error)
}

type IPlatform interface {
	Start() error
	Stop(signal string) error
	Kill(signal string) error
	Restart() error
	Delete() error
	Rename(newName string) error
	Exec(command []string) (types.ExecResult, error)
	Logs(bool) (io.ReadCloser, error)

	GetContainerState() (string, error)
	Run() (*TDTypes.Container, error)
	Prepare(config *configuration.Configuration, client *client.Http, user *authentication.User, runtime *types.Runtime) error
	PostRun(config *configuration.Configuration, dnsCache *dns.Records) error

	AttachToNetworks(string) error
	UpdateDns(dnsCache *dns.Records) error
	RemoveDns(dnsCache *dns.Records, networkId string) error
	SyncNetworkInformation() error
	GenerateLabels() map[string]string

	GetId() string
	GetDefinition() contracts.IDefinition
	GetGeneratedName() string
	GetName() string
	GetGroup() string
	GetGroupIdentifier() string
	GetDomain(networkName string) string
	GetHeadlessDomain(networkName string) string

	ToJson() ([]byte, error)
}

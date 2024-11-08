package platforms

import (
	TDTypes "github.com/docker/docker/api/types"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/container/status"
)

type IContainer interface {
	IsDaemonRunning()

	Start() bool
	Stop() bool
	Restart() bool
	Delete() error
	Rename(newName string) error
	Exec(command []string) types.ExecResult

	Get() (*TDTypes.Container, error)
	Run(*configuration.Environment, *client.Http, *dns.Records, *authentication.User) (*TDTypes.Container, error)
	Prepare(client *client.Http, user *authentication.User) error

	AttachToNetworks() error
	UpdateDns(dnsCache *dns.Records)

	GetRuntime() *types.Runtime
	GetStatus() *status.Status

	GetDefinition() v1.ContainerDefinition
	GetLabels() map[string]string
	GetGeneratedName() string
	GetName() string
	GetGroup() string
	GetGroupIdentifier() string

	GetDomain(network string) string
	GetHeadlessDomain(network string) string
}

type IPlatform interface {
	IsDaemonRunning()

	Start(runtime *types.Runtime) bool
	Stop(runtime *types.Runtime) bool
	Restart(runtime *types.Runtime) bool
	Delete(runtime *types.Runtime) error
	Rename(runtime *types.Runtime, newName string) error
	Exec(runtime *types.Runtime, command []string) types.ExecResult

	Get() (*TDTypes.Container, error)
	Run(*configuration.Environment, *client.Http, *dns.Records, *authentication.User) (*TDTypes.Container, error)
	Prepare(client *client.Http, user *authentication.User, runtime *types.Runtime) error

	AttachToNetworks() error
	UpdateDns(dnsCache *dns.Records)
	GenerateLabels() map[string]string

	GetDefinition() v1.ContainerDefinition
	GetGeneratedName() string
	GetName() string
	GetGroup() string
	GetGroupIdentifier() string
	GetDomain(networkName string) string
	GetHeadlessDomain(networkName string) string
}

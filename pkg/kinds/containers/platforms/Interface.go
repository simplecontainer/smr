package platforms

import (
	"bufio"
	"context"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/image"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/readiness"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/state"
	"github.com/simplecontainer/smr/pkg/kinds/containers/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/containers/status"
	"github.com/simplecontainer/smr/pkg/node"
	"io"
	"net"
)

type IContainer interface {
	Run() error
	PreRun(config *configuration.Configuration, client *clients.Http, user *authentication.User) error
	PostRun(config *configuration.Configuration, dnsCache *dns.Records) error
	InitContainer(definition *v1.ContainersInternal, config *configuration.Configuration, client *clients.Http, user *authentication.User) error
	MountResources() error

	UpdateDns(dnsCache *dns.Records) error
	RemoveDns(dnsCache *dns.Records, networkId string) error
	SyncNetwork() error

	HasDependencyOn(string, string, string) bool
	HasOwner() bool

	GetReadiness() []*readiness.Readiness
	GetState() (state.State, error)
	GetEngineState() string
	GetRuntime() *types.Runtime
	GetStatus() *status.Status
	GetNode() *node.Node
	GetNodeName() string
	GetId() string
	GetGlobalDefinition() *v1.ContainersDefinition
	GetDefinition() idefinitions.IDefinition
	GetLabels() map[string]string
	GetGeneratedName() string
	GetName() string
	GetGroup() string
	GetGroupIdentifier() string
	GetIndex() (uint64, error)
	GetImageState() *image.ImageState
	GetImageWithTag() string
	GetNetwork() map[string]net.IP
	GetDomain(network string) string
	GetHeadlessDomain(network string) string
	GetInit() IPlatform
	GetInitDefinition() *v1.ContainersInternal

	IsGhost() bool
	SetGhost(bool)

	CreateVolume(definition *v1.VolumeDefinition) error
	DeleteVolume(id string, force bool) error

	Start() error
	Stop(signal string) error
	Kill(signal string) error
	Restart() error
	Delete() error
	Wait(string) error
	Rename(newName string) error
	Exec(ctx context.Context, command []string, interactive bool, height string, width string) (string, *bufio.Reader, net.Conn, error)
	ExecInspect(string) (bool, int, error)
	ExecResize(string, int, int) error
	Logs(context.Context, bool) (io.ReadCloser, error)
	Clean() error

	ToJSON() ([]byte, error)
}

type IPlatform interface {
	Run() error
	PreRun(config *configuration.Configuration, client *clients.Http, user *authentication.User, runtime *types.Runtime) error
	PostRun(config *configuration.Configuration, dnsCache *dns.Records) error
	InitContainer(definition *v1.ContainersInternal, config *configuration.Configuration, client *clients.Http, user *authentication.User, runtime *types.Runtime) error
	MountResources() error

	UpdateDns(dnsCache *dns.Records) error
	RemoveDns(dnsCache *dns.Records, networkId string) error
	SyncNetwork() error

	GetReadiness() []*readiness.Readiness

	GetState() (state.State, error)
	GetEngineState() string
	GetId() string
	GetDefinition() idefinitions.IDefinition
	GetGeneratedName() string
	GetName() string
	GetGroup() string
	GetIndex() (uint64, error)
	GetGroupIdentifier() string
	GetImageState() *image.ImageState
	GetImageWithTag() string
	GetNetwork() map[string]net.IP
	GetDomain(networkName string) string
	GetHeadlessDomain(networkName string) string
	GetInit() IPlatform
	GetInitDefinition() *v1.ContainersInternal

	CreateVolume(definition *v1.VolumeDefinition) error
	DeleteVolume(id string, force bool) error

	Start() error
	Stop(signal string) error
	Kill(signal string) error
	Restart() error
	Delete() error
	Rename(newName string) error
	Exec(ctx context.Context, command []string, interactive bool, height string, width string) (string, *bufio.Reader, net.Conn, error)
	ExecInspect(string) (bool, int, error)
	ExecResize(string, int, int) error
	Logs(context.Context, bool) (io.ReadCloser, error)
	Wait(string) error
	Clean() error

	ToJSON() ([]byte, error)
}

type Registry interface {
	AddOrUpdate(group string, name string, containerAddr IContainer)
	Sync(group string, name string) error

	Remove(prefix string, group string, name string) error
	FindLocal(group string, name string) IContainer
	Find(prefix string, group string, name string) IContainer
	FindGroup(prefix string, group string) []IContainer
	Name(client *clients.Http, prefix string, group string, name string) (string, []uint64, error)
	NameReplica(group string, name string, index uint64) string
	BackOff(group string, name string) error
	BackOffReset(group string, name string)
	GetIndexes(prefix string, group string, name string) ([]uint64, error)
	GetIndexesLocal(prefix string, group string, name string) ([]uint64, error)
}

type Readiness interface {
}

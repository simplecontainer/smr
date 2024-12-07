package platforms

import (
	"errors"
	TDTypes "github.com/docker/docker/api/types"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/container/status"
	"github.com/simplecontainer/smr/pkg/static"
	"strconv"
	"time"
)

func New(platform string, name string, config *configuration.Configuration, definition *v1.ContainerDefinition) (IContainer, error) {
	statusObj := &status.Status{
		State:      &status.StatusState{},
		LastUpdate: time.Now(),
	}

	statusObj.CreateGraph()

	switch platform {
	case static.PLATFORM_DOCKER:
		dockerPlatform, err := docker.New(name, config, definition)

		if err != nil {
			return nil, err
		}

		return Container{
			Platform: dockerPlatform,
			General: &General{
				Runtime: &types.Runtime{
					Configuration:      make(map[string]string),
					ObjectDependencies: make([]*f.Format, 0),
					NodeIP:             strconv.FormatUint(config.KVStore.Node, 10),
					Agent:              config.Agent,
				},
				Status: statusObj,
			},
			Type: static.PLATFORM_DOCKER,
		}, nil
	default:
		return nil, errors.New("container platform is not implemented")
	}
}

func (c Container) Start() bool {
	return c.Platform.Start()
}
func (c Container) Stop() bool {
	return c.Platform.Stop()
}
func (c Container) Restart() bool {
	return c.Platform.Restart()
}
func (c Container) Delete() error {
	return c.Platform.Delete()
}
func (c Container) Rename(newName string) error {
	return c.Platform.Rename(newName)
}
func (c Container) Exec(command []string) types.ExecResult {
	return c.Platform.Exec(command)
}

func (c Container) Get() (*TDTypes.Container, error) {
	return c.Platform.Get()
}
func (c Container) Run(environment *configuration.Environment, http *client.Http, records *dns.Records, user *authentication.User) (*TDTypes.Container, error) {
	return c.Platform.Run(environment, http, records, user)
}
func (c Container) Prepare(client *client.Http, user *authentication.User) error {
	return c.Platform.Prepare(client, user, c.General.Runtime)
}

func (c Container) AttachToNetworks() error {
	return c.Platform.AttachToNetworks()
}
func (c Container) UpdateDns(cache *dns.Records) {
	c.Platform.UpdateDns(cache)
}

func (c Container) GetRuntime() *types.Runtime {
	return c.General.Runtime
}
func (c Container) GetStatus() *status.Status {
	return c.General.Status
}

func (c Container) GetDefinition() v1.ContainerDefinition {
	return c.Platform.GetDefinition()
}
func (c Container) GetLabels() map[string]string {
	return c.General.Labels
}
func (c Container) GetGeneratedName() string {
	return c.Platform.GetGeneratedName()
}
func (c Container) GetName() string {
	return c.Platform.GetName()
}
func (c Container) GetGroup() string {
	return c.Platform.GetGroup()
}
func (c Container) GetGroupIdentifier() string {
	return c.Platform.GetGroupIdentifier()
}

func (c Container) GetDomain(network string) string {
	return c.Platform.GetDomain(network)
}
func (c Container) GetHeadlessDomain(network string) string {
	return c.Platform.GetHeadlessDomain(network)
}

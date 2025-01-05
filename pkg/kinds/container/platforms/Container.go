package platforms

import (
	"encoding/json"
	"errors"
	TDTypes "github.com/docker/docker/api/types"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/container/distributed"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/container/status"
	"github.com/simplecontainer/smr/pkg/static"
	"strconv"
)

func New(platform string, name string, config *configuration.Configuration, ChangeC chan distributed.Container, definition *v1.ContainerDefinition) (IContainer, error) {
	statusObj := status.New(ChangeC)

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
					Agent:              config.Node,
				},
				Status: statusObj,
			},
			Type: static.PLATFORM_DOCKER,
		}, nil
	default:
		return nil, errors.New("container platform is not implemented")
	}
}

func NewGhost(state map[string]interface{}) (IContainer, error) {
	if state["Type"] != nil {
		switch state["Type"].(string) {
		case static.PLATFORM_DOCKER:
			ghost := &Container{
				Platform: &docker.Docker{},
				General:  &General{},
				Type:     static.PLATFORM_DOCKER,
				ghost:    true,
			}

			bytes, err := json.Marshal(state)

			if err != nil {
				return nil, err
			}

			err = json.Unmarshal(bytes, &ghost)
			if err != nil {
				return nil, err
			}

			return ghost, nil
		default:
			return nil, errors.New("container platform is not implemented")
		}
	}

	return nil, errors.New("type is not defined")
}

func (c Container) Start() error {
	return c.Platform.Start()
}
func (c Container) Stop(signal string) error {
	return c.Platform.Stop(signal)
}
func (c Container) Kill(signal string) error {
	return c.Platform.Kill(signal)
}
func (c Container) Restart() error {
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
func (c Container) Run(config *configuration.Configuration, http *client.Http, records *dns.Records, user *authentication.User) (*TDTypes.Container, error) {
	return c.Platform.Run(config, http, records, user)
}
func (c Container) Prepare(client *client.Http, user *authentication.User) error {
	return c.Platform.Prepare(client, user, c.General.Runtime)
}

func (c Container) AttachToNetworks(agentContainerName string) error {
	return c.Platform.AttachToNetworks(agentContainerName)
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
func (c Container) GetAgent() string {
	return c.General.Runtime.Agent
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

func (c Container) SetGhost(ghost bool) {
	c.ghost = ghost
}
func (c Container) IsGhost() bool {
	return c.ghost
}

func (c Container) ToJson() ([]byte, error) {
	return json.Marshal(c)
}

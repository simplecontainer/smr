package platforms

import (
	"encoding/json"
	"errors"
	TDTypes "github.com/docker/docker/api/types"
	jsoniter "github.com/json-iterator/go"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/engines/docker"
	"github.com/simplecontainer/smr/pkg/kinds/container/platforms/types"
	"github.com/simplecontainer/smr/pkg/kinds/container/status"
	"github.com/simplecontainer/smr/pkg/smaps"
	"github.com/simplecontainer/smr/pkg/static"
	"io"
)

func New(platform string, name string, config *configuration.Configuration, definition contracts.IDefinition) (IContainer, error) {
	statusObj := status.New()

	switch platform {
	case static.PLATFORM_DOCKER:
		dockerPlatform, err := docker.New(name, definition)

		if err != nil {
			return nil, err
		}

		return &Container{
			Platform: dockerPlatform,
			General: &General{
				Runtime: &types.Runtime{
					Configuration:      smaps.New(),
					ObjectDependencies: make([]f.Format, 0),
					NodeURL:            config.KVStore.URL,
					Node:               config.KVStore.Node,
					NodeName:           config.NodeName,
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

			var json = jsoniter.ConfigCompatibleWithStandardLibrary
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

func (c *Container) Start() error {
	return c.Platform.Start()
}
func (c *Container) Stop(signal string) error {
	return c.Platform.Stop(signal)
}
func (c *Container) Kill(signal string) error {
	return c.Platform.Kill(signal)
}
func (c *Container) Restart() error {
	return c.Platform.Restart()
}
func (c *Container) Delete() error {
	return c.Platform.Delete()
}
func (c *Container) Rename(newName string) error {
	return c.Platform.Rename(newName)
}
func (c *Container) Exec(command []string) (types.ExecResult, error) {
	return c.Platform.Exec(command)
}
func (c *Container) Logs(follow bool) (io.ReadCloser, error) {
	return c.Platform.Logs(follow)
}

func (c *Container) GetContainerState() (string, error) {
	return c.Platform.GetContainerState()
}
func (c *Container) Run() (*TDTypes.Container, error) {
	return c.Platform.Run()
}
func (c *Container) Prepare(config *configuration.Configuration, client *client.Http, user *authentication.User) error {
	return c.Platform.Prepare(config, client, user, c.General.Runtime)
}
func (c *Container) PostRun(config *configuration.Configuration, dnsCache *dns.Records) error {
	return c.Platform.PostRun(config, dnsCache)
}

func (c *Container) AttachToNetworks(agentContainerName string) error {
	return c.Platform.AttachToNetworks(agentContainerName)
}
func (c *Container) UpdateDns(cache *dns.Records) error {
	return c.Platform.UpdateDns(cache)
}
func (c *Container) RemoveDns(cache *dns.Records, networkId string) error {
	return c.Platform.RemoveDns(cache, networkId)
}
func (c *Container) SyncNetworkInformation() error {
	return c.Platform.SyncNetworkInformation()
}

func (c *Container) HasDependencyOn(kind string, group string, name string) bool {
	for _, format := range c.GetRuntime().ObjectDependencies {
		if format.Name == name && format.Group == group && format.Kind == kind {
			return true
		}
	}

	return false
}

func (c *Container) HasOwner() bool {
	return c.GetDefinition().GetRuntime().GetOwner().IsEmpty()
}

func (c *Container) GetId() string {
	return c.GetId()
}

func (c *Container) GetRuntime() *types.Runtime {
	return c.General.Runtime
}
func (c *Container) GetStatus() *status.Status {
	return c.General.Status
}
func (c *Container) GetNode() uint64 {
	return c.General.Runtime.Node
}

func (c *Container) GetNodeName() string {
	return c.General.Runtime.NodeName
}

func (c *Container) GetDefinition() contracts.IDefinition {
	return c.Platform.GetDefinition()
}
func (c *Container) GetLabels() map[string]string {
	return c.General.Labels
}
func (c *Container) GetGeneratedName() string {
	return c.Platform.GetGeneratedName()
}
func (c *Container) GetName() string {
	return c.Platform.GetName()
}
func (c *Container) GetGroup() string {
	return c.Platform.GetGroup()
}
func (c *Container) GetGroupIdentifier() string {
	return c.Platform.GetGroupIdentifier()
}

func (c *Container) GetDomain(network string) string {
	return c.Platform.GetDomain(network)
}
func (c *Container) GetHeadlessDomain(network string) string {
	return c.Platform.GetHeadlessDomain(network)
}

func (c *Container) SetGhost(ghost bool) {
	c.ghost = ghost
}
func (c *Container) IsGhost() bool {
	return c.ghost
}

func (c *Container) ToJson() ([]byte, error) {
	var output = make(map[string]json.RawMessage)
	var err error

	output["Platform"], err = c.Platform.ToJson()

	if err != nil {
		return nil, err
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	output["General"], err = json.Marshal(c.General)

	if err != nil {
		return nil, err
	}

	output["Type"], err = json.Marshal(c.Type)

	if err != nil {
		return nil, err
	}

	return json.Marshal(output)
}

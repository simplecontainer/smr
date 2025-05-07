package controls

import (
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/manager"
	"time"
)

func NewCommand(name string, data map[string]string) *GenericCommand {
	return &GenericCommand{
		name:      name,
		data:      data,
		timestamp: time.Now(),
	}
}

func (c *GenericCommand) Name() string {
	return c.name
}

func (c *GenericCommand) Node(mgr *manager.Manager, params map[string]string) error {
	return c.Node(mgr, params)
}

func (c *GenericCommand) Agent(client *client.Client, params map[string]string) error {
	return c.Agent(client, params)
}

func (c *GenericCommand) Data() map[string]string {
	return c.data
}

func (c *GenericCommand) SetData(data map[string]string) {
	c.data = data
}

func (c *GenericCommand) NodeID() uint64 {
	return c.nodeID
}

func (c *GenericCommand) SetNodeID(id uint64) {
	c.nodeID = id
}

func (c *GenericCommand) Time() time.Time {
	return c.timestamp
}

func (c *GenericCommand) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name      string            `json:"name"`
		Data      map[string]string `json:"data"`
		Timestamp time.Time         `json:"timestamp"`
	}{
		Name:      c.name,
		Data:      c.data,
		Timestamp: c.timestamp,
	})
}

func (c *GenericCommand) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Name      string            `json:"name"`
		Data      map[string]string `json:"data"`
		Timestamp time.Time         `json:"timestamp"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	c.name = aux.Name
	c.data = aux.Data
	c.timestamp = aux.Timestamp

	return nil
}

func (c *GenericCommand) String() string {
	return fmt.Sprintf("Command{name=%s}", c.name)
}

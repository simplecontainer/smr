package controler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func New() *Control {
	return &Control{
		Drain:   &Drain{},
		Upgrade: &Upgrade{},
		Start:   &Start{},
	}
}

func (c *Control) SetStart(start *Start) {
	c.Start = start
}

func (c *Control) SetDrain(drain *Drain) {
	c.Drain = drain
}

func (c *Control) SetUpgrade(upgrade *Upgrade) {
	c.Upgrade = upgrade
}

func (c *Control) GetStart() *Start {
	return c.Start
}

func (c *Control) GetDrain() *Drain {
	return c.Drain
}

func (c *Control) GetUpgrade() *Upgrade {
	return c.Upgrade
}

func NewUpgrade(image string, tag string) *Upgrade {
	return &Upgrade{
		Image: image,
		Tag:   tag,
	}
}
func NewDrain(nodeID uint64) *Drain {
	return &Drain{
		NodeID: nodeID,
	}
}
func NewStart(nodeAPI string, overlay string, backend string) *Start {
	return &Start{
		NodeAPI: nodeAPI,
		Overlay: overlay,
		Backend: backend,
	}
}

func (c *Control) Apply(ctx context.Context, client *clientv3.Client) error {
	bytes, err := c.ToJSON()

	if err != nil {
		return err
	}

	_, err = client.Put(ctx, "/smr/control", string(bytes))
	return err
}
func (c *Control) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

func (c *Control) Validate() (bool, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	if c.Start != nil {
		if err := validate.Struct(c.Start); err != nil {
			return false, fmt.Errorf("start validation failed: %w", err)
		}
	}

	if c.Drain != nil {
		if err := validate.Struct(c.Drain); err != nil {
			return false, fmt.Errorf("drain validation failed: %w", err)
		}
	}

	if c.Upgrade != nil {
		if err := validate.Struct(c.Upgrade); err != nil {
			return false, fmt.Errorf("upgrade validation failed: %w", err)
		}
	}

	if c.Start == nil && c.Drain == nil && c.Upgrade == nil {
		return false, fmt.Errorf("control is empty")
	}

	return true, nil
}

package controler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func New() *Control {
	return &Control{
		Drain:   &Drain{},
		Upgrade: &Upgrade{},
	}
}

func (c *Control) SetDrain(drain *Drain) {
	c.Drain = drain
}

func (c *Control) SetUpgrade(upgrade *Upgrade) {
	c.Upgrade = upgrade
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

	err := validate.Struct(c)
	if err != nil {
		var invalidValidationError *validator.InvalidValidationError
		if errors.As(err, &invalidValidationError) {
			return false, err
		}
		// from here you can create your own error messages in whatever language you wish
		return false, err
	}

	return true, nil
}

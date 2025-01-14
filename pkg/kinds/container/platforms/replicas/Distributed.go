package replicas

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/objects"
)

func NewDistributed(nodeID uint64, group string, name string, replicas *Replicas) *Distributed {
	dr := &Distributed{
		Group:    group,
		Name:     name,
		Replicas: make(map[uint64]*Replicas),
	}

	dr.Replicas[nodeID] = replicas
	return dr
}

func (dr *Distributed) Save(client *client.Client, user *authentication.User) error {
	format := f.NewFromString(fmt.Sprintf("replicas.%s.%s", dr.Group, dr.Name))
	obj := objects.New(client, user)

	data, err := dr.ToJson()

	if err != nil {
		return err
	}

	obj.Update(format, data)
	return nil
}

func (dr *Distributed) Remove(client *client.Client, user *authentication.User) (bool, error) {
	format := f.NewFromString(fmt.Sprintf("replicas.%s.%s", dr.Group, dr.Name))
	obj := objects.New(client, user)

	return obj.Remove(format)
}

func (dr *Distributed) Load(client *client.Client, user *authentication.User) error {
	format := f.NewFromString(fmt.Sprintf("replicas.%s.%s", dr.Group, dr.Name))
	obj := objects.New(client, user)

	obj.Find(format)

	if obj.Exists() {
		err := dr.FromJson(obj.GetDefinitionByte())

		if err != nil {
			return err
		}

		return nil
	} else {
		return errors.New("distributed replicas not found")
	}
}

func (dr *Distributed) ToJson() ([]byte, error) {
	return json.Marshal(dr)
}

func (dr *Distributed) FromJson(data []byte) error {
	return json.Unmarshal(data, dr)
}

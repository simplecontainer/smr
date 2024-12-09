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

func NewDistributed(nodeID uint64, group string, name string) *DistributedReplicas {
	dr := &DistributedReplicas{
		Group:    group,
		Name:     name,
		Replicas: make(map[uint64]*ScopedReplicas),
	}

	dr.Replicas[nodeID] = &ScopedReplicas{
		Create: make([]R, 0),
		Remove: make([]R, 0),
		Numbers: Numbers{
			Create:   make([]uint64, 0),
			Destroy:  make([]uint64, 0),
			Existing: make([]uint64, 0),
		},
	}

	return dr
}

func (dr *DistributedReplicas) Save(client *client.Client, user *authentication.User) error {
	format := f.NewFromString(fmt.Sprintf("replicas.%s.%s", dr.Group, dr.Name))
	obj := objects.New(client, user)

	data, err := dr.ToJson()

	if err != nil {
		return err
	}

	obj.Update(format, string(data))
	return nil
}

func (dr *DistributedReplicas) Clear(node uint64) {
	dr.Replicas[node].Create = make([]R, 0)
	dr.Replicas[node].Remove = make([]R, 0)
}

func (dr *DistributedReplicas) Load(client *client.Client, user *authentication.User) error {
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

func (dr *DistributedReplicas) ToJson() ([]byte, error) {
	return json.Marshal(dr)
}

func (dr *DistributedReplicas) FromJson(data []byte) error {
	return json.Unmarshal(data, dr)
}

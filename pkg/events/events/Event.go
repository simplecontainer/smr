package events

import (
	"encoding/json"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/raft"
	"github.com/simplecontainer/smr/pkg/static"
)

func New(event string, target string, kind string, group string, name string, data []byte) Event {
	return Event{
		Type:   event,
		Target: target,
		Kind:   kind,
		Group:  group,
		Name:   name,
		Data:   data,
	}
}

func (event Event) Propose(proposeC *raft.KVStore, node uint64) error {
	bytes, err := json.Marshal(event)

	if err != nil {
		return err
	}

	proposeC.Propose(event.GetKey(), bytes, node)
	return nil
}

func (event Event) GetType() string {
	return event.Type
}

func (event Event) GetTarget() string {
	return event.Target

}

func (event Event) GetKind() string {
	return event.Kind
}

func (event Event) GetGroup() string {
	return event.Group

}

func (event Event) GetName() string {
	return event.Name

}

func (event Event) GetData() []byte {
	return event.Data
}

func (event Event) GetKey() string {
	format := f.New(static.SMR_PREFIX, static.CATEGORY_EVENT, event.Kind, event.Group, event.Name)
	return format.ToString()
}

func (event Event) GetNetworkId() string {
	return ""
}

func (event Event) GetContainerId() string {
	return ""
}

func (event Event) IsManaged() bool {
	return true
}

func (event Event) ToJson() ([]byte, error) {
	return json.Marshal(event)
}

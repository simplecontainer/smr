package events

import (
	"encoding/json"
	"github.com/simplecontainer/smr/pkg/contracts/idefinitions"
	"github.com/simplecontainer/smr/pkg/contracts/iformat"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/raft"
	"github.com/simplecontainer/smr/pkg/static"
)

func New(event string, target string, prefix string, kind string, group string, name string, data []byte) Event {
	return Event{
		Type:   event,
		Target: target,
		Prefix: prefix,
		Kind:   kind,
		Group:  group,
		Name:   name,
		Data:   data,
	}
}

func NewNodeEvent(event string, n *node.Node) (Event, error) {
	bytes, err := n.ToJSON()

	if err != nil {
		return Event{}, err
	}

	return Event{
		Type:   event,
		Target: static.KIND_NODE,
		Prefix: static.SMR_PREFIX,
		Kind:   static.KIND_NODE,
		Group:  "internal",
		Name:   n.NodeName,
		Data:   bytes,
	}, nil
}

func NewKindEvent(event string, definition idefinitions.IDefinition, data []byte) Event {
	switch event {
	case EVENT_INSPECT:
		if definition.GetRuntime().GetOwner().IsEmpty() {
			return Event{}
		} else {
			return Event{
				Type:   event,
				Target: definition.GetRuntime().GetOwner().Kind,
				Prefix: definition.GetPrefix(),
				Kind:   definition.GetRuntime().GetOwner().Kind,
				Group:  definition.GetRuntime().GetOwner().Group,
				Name:   definition.GetRuntime().GetOwner().Name,
				Data:   data,
			}
		}
	default:
		return Event{
			Type:   event,
			Target: definition.GetKind(),
			Prefix: definition.GetPrefix(),
			Kind:   definition.GetKind(),
			Group:  definition.GetMeta().Group,
			Name:   definition.GetMeta().Name,
			Data:   data,
		}
	}
}

func (e Event) SetName(name string) Event {
	e.Name = name
	return e
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

func (event Event) GetPrefix() string {
	return event.Prefix
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

func (event Event) ToFormat() f.Format {
	return f.New(event.GetPrefix(), event.GetKind(), event.GetGroup(), event.GetName())
}

func (event Event) IsOfFormat(format iformat.Format) bool {
	opts := f.DefaultToStringOpts()
	opts.ExcludeCategory = true

	if event.ToFormat().ToString() == format.ToStringWithOpts(opts) {
		return true
	} else {
		return false
	}
}

func (event Event) ToJSON() ([]byte, error) {
	return json.Marshal(event)
}

func (event Event) IsEmpty() bool {
	return event.Kind == ""
}

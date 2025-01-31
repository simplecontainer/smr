package events

import (
	"encoding/json"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/static"
)

func New(event string, kind string, group string, name string, data []byte) Events {
	return Events{
		Type:  event,
		Kind:  kind,
		Group: group,
		Name:  name,
		Data:  data,
	}
}

func (event *Events) GetKey() string {
	return f.New(static.SMR_PREFIX, static.CATEGORY_EVENT, event.Kind, event.Group, event.Name).ToString()
}

func (event *Events) ToJson() ([]byte, error) {
	return json.Marshal(event)
}

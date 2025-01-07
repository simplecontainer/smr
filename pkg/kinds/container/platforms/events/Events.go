package events

import (
	"encoding/json"
	"fmt"
)

func New(kind string, group string, name string, data []byte) Events {
	return Events{
		Kind:  kind,
		Group: group,
		Name:  name,
		Data:  data,
	}
}

func (event *Events) GetKey() string {
	return fmt.Sprintf("%s.%s", event.Group, event.Name)
}

func (event *Events) ToJson() ([]byte, error) {
	return json.Marshal(event)
}

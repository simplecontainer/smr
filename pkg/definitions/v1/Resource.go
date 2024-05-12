package v1

import "encoding/json"

type Resource struct {
	Meta ResourceMeta `json:"meta"`
	Spec ResourceSpec `json:"spec"`
}

type ResourceMeta struct {
	Group      string `json:"group"`
	Identifier string `json:"identifier"`
}

type ResourceSpec struct {
	Data map[string]any `json:"data"`
}

func (resource *Resource) ToJsonString() (string, error) {
	bytes, err := json.Marshal(resource)
	return string(bytes), err
}

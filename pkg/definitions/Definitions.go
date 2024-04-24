package definitions

import (
	"encoding/json"
)

func (definition *Container) ToJsonString() (string, error) {
	bytes, err := json.Marshal(definition)
	return string(bytes), err
}

func (configuration *Configuration) ToJsonString() (string, error) {
	bytes, err := json.Marshal(configuration)
	return string(bytes), err
}

func (resource *Resource) ToJsonString() (string, error) {
	bytes, err := json.Marshal(resource)
	return string(bytes), err
}

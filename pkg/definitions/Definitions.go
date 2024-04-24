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

func (operator *Operator) ToJsonString() (string, error) {
	bytes, err := json.Marshal(operator)
	return string(bytes), err
}

func (resource *Resource) ToJsonString() (string, error) {
	bytes, err := json.Marshal(resource)
	return string(bytes), err
}

func (certkey *CertKey) ToJsonString() (string, error) {
	bytes, err := json.Marshal(certkey)
	return string(bytes), err
}

func (httpauth *HttpAuth) ToJsonString() (string, error) {
	bytes, err := json.Marshal(httpauth)
	return string(bytes), err
}

func (gitops *Gitops) ToJsonString() (string, error) {
	bytes, err := json.Marshal(gitops)
	return string(bytes), err
}

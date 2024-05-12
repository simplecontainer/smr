package v1

import "encoding/json"

type HttpAuth struct {
	Meta HttpAuthMeta `json:"meta"`
	Spec HttpAuthSpec `json:"spec"`
}

type HttpAuthMeta struct {
	Group      string `json:"group"`
	Identifier string `json:"identifier"`
}

type HttpAuthSpec struct {
	Username string
	Password string
}

func (httpauth *HttpAuth) ToJsonString() (string, error) {
	bytes, err := json.Marshal(httpauth)
	return string(bytes), err
}

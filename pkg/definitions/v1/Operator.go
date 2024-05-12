package v1

import "encoding/json"

type Operator struct {
	Meta Meta `json:"meta"`
	Spec Spec `json:"spec"`
}

func (operator *Operator) ToJsonString() (string, error) {
	bytes, err := json.Marshal(operator)
	return string(bytes), err
}

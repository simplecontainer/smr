package v1

import "github.com/simplecontainer/smr/pkg/definitions/commonv1"

type CustomDefinition struct {
	Meta CustomMeta `json:"meta"  validate:"required"`
	Spec CustomSpec `json:"spec"  validate:"required"`
}

type CustomMeta struct {
	Name    string            `validate:"required" json:"name"`
	Group   string            `validate:"required" json:"group"`
	Runtime *commonv1.Runtime `json:"runtime"`
}

type CustomSpec struct {
	Custom CustomInternal `validate:"required" json:"custom" `
}

type CustomInternal struct {
	Definition []byte
}

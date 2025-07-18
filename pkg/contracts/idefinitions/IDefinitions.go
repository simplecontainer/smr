package idefinitions

import (
	"github.com/simplecontainer/smr/pkg/contracts/iobjects"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
)

type IDefinition interface {
	FromJson([]byte) error
	SetRuntime(*commonv1.Runtime)
	GetRuntime() *commonv1.Runtime
	GetPrefix() string
	GetMeta() *commonv1.Meta
	GetState() *commonv1.State
	SetState(*commonv1.State)
	GetKind() string
	ResolveReferences(iobjects.ObjectInterface) ([]IDefinition, error)
	ToJSON() ([]byte, error)
	ToYAML() ([]byte, error)
	ToJSONString() (string, error)
	Validate() (bool, error)
}

package contracts

import (
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
)

type IDefinition interface {
	FromJson([]byte) error
	SetRuntime(*commonv1.Runtime)
	GetRuntime() *commonv1.Runtime
	GetMeta() commonv1.Meta
	GetKind() string
	ResolveReferences(ObjectInterface) ([]IDefinition, error)
	ToJson() ([]byte, error)
	ToJsonWithKind() ([]byte, error)
	ToJsonString() (string, error)
	Validate() (bool, error)
}

type TDefinition interface {
	Apply(Format, ObjectInterface, string) (ObjectInterface, error)
	Delete(Format, ObjectInterface, string) (IDefinition, error)
	Changed(Format, ObjectInterface) (bool, error)
	FromJson([]byte) error
	SetRuntime(*commonv1.Runtime)
	GetRuntime() *commonv1.Runtime
	GetMeta() commonv1.Meta
	GetKind() string
	ResolveReferences(ObjectInterface) ([]IDefinition, error)
	ToJson() ([]byte, error)
	ToJsonWithKind() ([]byte, error)
	ToJsonString() (string, error)
	Validate() (bool, error)
}

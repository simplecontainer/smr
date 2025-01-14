package contracts

import (
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/f"
)

type IDefinition interface {
	FromJson([]byte) error
	SetRuntime(*commonv1.Runtime)
	GetRuntime() *commonv1.Runtime
	GetKind() string
	ResolveReferences(ObjectInterface) ([]IDefinition, error)
	ToJson() ([]byte, error)
	ToJsonWithKind() ([]byte, error)
	ToJsonString() (string, error)
	Validate() (bool, error)
}

type TDefinition interface {
	Apply(*f.Format, ObjectInterface, string) (ObjectInterface, error)
	Delete(*f.Format, ObjectInterface, string) (IDefinition, error)
	Changed(*f.Format, ObjectInterface) (bool, error)
	FromJson([]byte) error
	SetRuntime(*commonv1.Runtime)
	GetRuntime() *commonv1.Runtime
	GetKind() string
	ResolveReferences(ObjectInterface) ([]IDefinition, error)
	ToJson() ([]byte, error)
	ToJsonWithKind() ([]byte, error)
	ToJsonString() (string, error)
	Validate() (bool, error)
}

package contracts

import (
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/f"
)

type IDefinition interface {
	FromJson([]byte) error
	SetOwner(string, string, string)
	GetOwner() commonv1.Owner
	GetKind() string
	ResolveReferences(ObjectInterface) ([]IDefinition, error)
	ToJson() ([]byte, error)
	ToJsonString() (string, error)
	ToJsonStringWithKind() (string, error)
	Validate() (bool, error)
}

type TDefinition interface {
	Apply(*f.Format, ObjectInterface, string) (ObjectInterface, error)
	Delete(*f.Format, ObjectInterface, string) (IDefinition, error)
	Changed(*f.Format, ObjectInterface) (bool, error)
	FromJson([]byte) error
	SetOwner(string, string, string)
	GetOwner() commonv1.Owner
	GetKind() string
	ResolveReferences(ObjectInterface) ([]IDefinition, error)
	ToJson() ([]byte, error)
	ToJsonString() (string, error)
	ToJsonStringWithKind() (string, error)
	Validate() (bool, error)
}

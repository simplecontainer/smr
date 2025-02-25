package idefinitions

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts/iobjects"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
)

type IDefinition interface {
	FromJson([]byte) error
	SetRuntime(*commonv1.Runtime)
	GetRuntime() *commonv1.Runtime
	GetPrefix() string
	GetMeta() commonv1.Meta
	GetState() *commonv1.State
	SetState(*commonv1.State)
	GetKind() string
	ResolveReferences(iobjects.ObjectInterface) ([]IDefinition, error)
	ToJson() ([]byte, error)
	ToJsonString() (string, error)
	Validate() (bool, error)
}

type TDefinition interface {
	Apply(client *client.Http, user *authentication.User) (iobjects.ObjectInterface, error)
	Delete(client *client.Http, user *authentication.User) (IDefinition, error)
	Changed(client *client.Http, user *authentication.User) (bool, error)
	ProposeApply(client *client.Http, user *authentication.User) (bool, error)
	ProposeRemove(client *client.Http, user *authentication.User) (bool, error)
	FromJson([]byte) error
	SetRuntime(*commonv1.Runtime)
	GetRuntime() *commonv1.Runtime
	GetPrefix() string
	GetMeta() commonv1.Meta
	GetState() *commonv1.State
	SetState(*commonv1.State)
	GetKind() string
	ResolveReferences(iobjects.ObjectInterface) ([]IDefinition, error)
	ToJson() ([]byte, error)
	ToJsonString() (string, error)
	Validate() (bool, error)
}

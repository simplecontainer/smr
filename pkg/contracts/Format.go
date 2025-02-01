package contracts

import "github.com/google/uuid"

type Format interface {
	GetPrefix() string
	GetVersion() string
	GetCategory() string
	GetType() string
	GetKind() string
	GetGroup() string
	GetName() string
	GetUUID() uuid.UUID
	ToString() string
	ToStringWithUUID() string
	Full() bool
	IsValid() bool
	ToBytes() []byte
}

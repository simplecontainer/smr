package iformat

import "github.com/google/uuid"

type Format interface {
	GetPrefix() string
	GetVersion() string
	GetCategory() string
	GetType() string
	GetKind() string
	GetGroup() string
	GetName() string
	Inverse() Format
	GetUUID() uuid.UUID
	ToString() string
	ToStringWithUUID() string
	Compliant() bool
	IsValid() bool
	ToBytes() []byte
}

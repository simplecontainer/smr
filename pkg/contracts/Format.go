package contracts

import "github.com/google/uuid"

type Format interface {
	GetCategory() string
	GetType() string
	GetUUID() uuid.UUID
	ToString() string
	ToStringWithUUID() string
	Full() bool
	WithPrefix(prefix string) Format
	IsValid() bool
	ToBytes() []byte
}

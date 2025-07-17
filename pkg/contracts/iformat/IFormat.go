package iformat

import (
	"github.com/google/uuid"
)

type Format interface {
	GetPrefix() string
	GetVersion() string
	GetCategory() string
	GetType() string
	GetKind() string
	GetGroup() string
	GetName() string
	GetField() string
	Shift() Format
	GetUUID() uuid.UUID
	ToString() string
	ToStringWithOpts(opts *ToStringOpts) string
	ToStringWithUUID() string
	Compliant() bool
	IsValid() bool
	ToBytes() []byte
}

type ToStringOpts struct {
	IncludeUUID      bool
	ExcludeCategory  bool
	AddTrailingSlash bool
	AddPrefixSlash   bool
}

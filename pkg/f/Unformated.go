package f

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/simplecontainer/smr/pkg/contracts/iformat"
	"strings"
)

func NewUnformated(key string) Unformated {
	UUID, f := parseUUID(key)

	return Unformated{
		Key:  strings.TrimPrefix(f, "/"),
		UUID: UUID,
		Type: TYPE_UNFORMATED,
	}
}

func (format Unformated) GetPrefix() string { return format.Prefix }

func (format Unformated) GetVersion() string { return format.Version }

func (format Unformated) GetCategory() string {
	return format.Category
}

func (format Unformated) GetType() string {
	return format.Type
}

func (format Unformated) GetKind() string {
	return ""
}

func (format Unformated) GetGroup() string {
	return ""
}

func (format Unformated) GetName() string {
	return ""
}

func (format Unformated) GetField() string {
	return ""
}

func (format Unformated) Shift() iformat.Format {
	return format
}

func (format Unformated) GetUUID() uuid.UUID {
	return format.UUID
}

func (format Unformated) IsValid() bool {
	return format.Key != ""
}

func (format Unformated) Compliant() bool {
	return format.Key != ""
}

func (format Unformated) ToString() string {
	return fmt.Sprintf("%s%s", format.Prefix, format.Key)
}

func (format Unformated) ToStringWithOpts(opts *iformat.ToStringOpts) string {
	return fmt.Sprintf("%s%s", format.Prefix, format.Key)
}

func (format Unformated) ToStringWithUUID() string {
	return fmt.Sprintf("%s%s", format.UUID, fmt.Sprintf("%s%s", format.Prefix, format.Key))
}

func (format Unformated) ToBytes() []byte {
	return []byte(fmt.Sprintf("%s%s", format.Prefix, format.Key))
}

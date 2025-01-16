package f

import (
	"fmt"
	"github.com/google/uuid"
)

func NewUnformated(key string, category string) Unformated {
	UUID, f := parseUUID(key)

	return Unformated{
		Key:      f,
		Category: category,
		UUID:     UUID,
		Type:     TYPE_UNFORMATED,
	}
}

func (format Unformated) GetCategory() string {
	return format.Category
}

func (format Unformated) GetType() string {
	return format.Type
}

func (format Unformated) GetUUID() uuid.UUID {
	return format.UUID
}

func (format Unformated) IsValid() bool {
	return format.Key != ""
}

func (format Unformated) Full() bool {
	return format.Key != ""
}

func (format Unformated) ToString() string {
	return format.Key
}

func (format Unformated) ToStringWithUUID() string {
	return fmt.Sprintf("%s%s", format.UUID, format.Key)
}

func (format Unformated) ToBytes() []byte {
	return []byte(format.Key)
}

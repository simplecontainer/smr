package contracts

import (
	"github.com/r3labs/diff/v3"
)

type ObjectInterface interface {
	GetDefinitionString() string
	GetDefinition() map[string]any
	GetDefinitionByte() []byte
	Add(format Format, data []byte) error
	AddLocal(format Format, data []byte) error
	Update(format Format, data []byte) error
	Find(format Format) error
	FindMany(format Format) (map[string]ObjectInterface, error)
	Remove(format Format) (bool, error)
	RemoveLocal(format Format) (bool, error)
	Diff(definition []byte) bool
	GetDiff() []diff.Change
	Exists() bool
	ChangeDetected() bool
}

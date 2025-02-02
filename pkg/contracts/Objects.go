package contracts

import (
	"github.com/r3labs/diff/v3"
)

type ObjectInterface interface {
	GetDefinition() map[string]any
	GetDefinitionByte() []byte
	Propose(Format, []byte) error
	Wait(Format, []byte) error
	AddLocal(format Format, data []byte) error
	RemoveLocal(format Format) (bool, error)
	AddLocalKey(key string, data []byte) error
	RemoveLocalKey(key string) (bool, error)
	Find(format Format) error
	FindMany(format Format) ([]ObjectInterface, error)
	Diff(definition []byte) bool
	GetDiff() []diff.Change
	Exists() bool
	ChangeDetected() bool
}

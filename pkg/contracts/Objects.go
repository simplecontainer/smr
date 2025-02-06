package contracts

import (
	"github.com/wI2L/jsondiff"
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
	GetDiff() jsondiff.Patch
	Exists() bool
	ChangeDetected() bool
}

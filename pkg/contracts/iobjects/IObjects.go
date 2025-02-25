package iobjects

import (
	"github.com/simplecontainer/smr/pkg/contracts/iformat"
	"github.com/wI2L/jsondiff"
)

type ObjectInterface interface {
	GetDefinition() map[string]any
	GetDefinitionByte() []byte
	Propose(iformat.Format, []byte) error
	Wait(iformat.Format, []byte) error
	AddLocal(format iformat.Format, data []byte) error
	RemoveLocal(format iformat.Format) (bool, error)
	AddLocalKey(key string, data []byte) error
	RemoveLocalKey(key string) (bool, error)
	Find(format iformat.Format) error
	FindMany(format iformat.Format) ([]ObjectInterface, error)
	Diff(definition []byte) bool
	GetDiff() jsondiff.Patch
	Exists() bool
	ChangeDetected() bool
}

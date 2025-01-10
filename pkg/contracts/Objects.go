package contracts

import (
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/pkg/f"
)

type ObjectInterface interface {
	GetDefinitionString() string
	GetDefinition() map[string]any
	GetDefinitionByte() []byte
	Add(format *f.Format, data []byte) error
	AddLocal(format *f.Format, data []byte) error
	Update(format *f.Format, data []byte) error
	Find(format *f.Format) error
	FindMany(format *f.Format) (map[string]ObjectInterface, error)
	Remove(format *f.Format) (bool, error)
	RemoveLocal(format *f.Format) (bool, error)
	Diff(definition []byte) bool
	GetDiff() []diff.Change
	Exists() bool
	ChangeDetected() bool
}

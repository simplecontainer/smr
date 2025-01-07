package objects

import (
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
	Remove(format *f.Format) (bool, error)
	Diff(definition []byte) bool
	Exists() bool
	ChangeDetected() bool
}

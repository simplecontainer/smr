package objects

import (
	"github.com/simplecontainer/smr/pkg/f"
)

type ObjectInterface interface {
	GetDefinitionString() string
	GetDefinition() map[string]any
	GetDefinitionByte() []byte
	Add(format *f.Format, data string) error
	Update(format *f.Format, data string) error
	Find(format *f.Format) error
	Remove(format *f.Format) (bool, error)
	Diff(definition string) bool
	Exists() bool
	ChangeDetected() bool
}

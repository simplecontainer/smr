package contracts

import (
	"github.com/google/uuid"
	"github.com/r3labs/diff/v3"
)

type ObjectInterface interface {
	GetDefinitionString() string
	GetDefinition() map[string]any
	GetDefinitionByte() []byte
	Propose(Format, []byte) (uuid.UUID, error)
	Wait(uuid uuid.UUID) error
	AddLocal(format Format, data []byte) error
	RemoveLocal(format Format) (bool, error)
	Find(format Format) error
	FindMany(format Format) (map[string]ObjectInterface, error)
	Diff(definition []byte) bool
	GetDiff() []diff.Change
	Exists() bool
	ChangeDetected() bool
}

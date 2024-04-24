package objects

import (
	"github.com/r3labs/diff/v3"
	"time"
)

type Object struct {
	Changelog      diff.Changelog
	definition     map[string]any
	definitionByte []byte
	changed        bool
	exists         bool
	created        time.Time
	updated        time.Time
}

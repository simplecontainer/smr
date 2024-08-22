package objects

import (
	"github.com/r3labs/diff/v3"
	"net/http"
	"time"
)

type Object struct {
	Changelog diff.Changelog
	client    *http.Client

	definition       map[string]any
	DefinitionString string
	definitionByte   []byte
	changed          bool
	exists           bool
	restoring        bool
	Created          time.Time
	Updated          time.Time
}

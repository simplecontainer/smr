package objects

import (
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"time"
)

type Object struct {
	Changelog diff.Changelog
	Raw       bool
	client    *client.Client

	definition       map[string]any
	DefinitionString string
	definitionByte   []byte
	changed          bool
	exists           bool
	restoring        bool
	Owner            string
	Created          time.Time
	Updated          time.Time
	User             *authentication.User
}

package secrets

import (
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"time"
)

type Object struct {
	Byte       []byte
	String     string
	Definition map[string]interface{}
	Category   string

	Changelog diff.Changelog
	Raw       bool
	client    *client.Client

	changed   bool
	exists    bool
	restoring bool
	Owner     string
	Created   time.Time
	Updated   time.Time
	User      *authentication.User
}

package objects

import (
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"time"
)

type Object struct {
	Byte     []byte
	Category string

	Changelog *diff.Changelog
	client    *client.Client

	changed bool
	exists  bool
	Created time.Time
	Updated time.Time
	User    *authentication.User
}

package objects

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/wI2L/jsondiff"
	"time"
)

type Object struct {
	Byte     []byte
	Category string

	Changelog jsondiff.Patch
	client    *client.Client

	changed bool
	exists  bool
	Created time.Time
	Updated time.Time
	User    *authentication.User
}

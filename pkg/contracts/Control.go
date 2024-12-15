package contracts

import (
	"github.com/simplecontainer/smr/pkg/authentication"
)

type Control struct {
	Kind      string
	Operation string
	Group     string
	Name      string

	User *authentication.User
	Data map[string]any
}

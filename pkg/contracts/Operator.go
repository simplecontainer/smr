package contracts

import (
	"github.com/simplecontainer/smr/pkg/authentication"
)

type ResponseOperator struct {
	HttpStatus       int
	Explanation      string
	ErrorExplanation string
	Error            bool
	Success          bool
	Data             map[string]any
}

type RequestOperator struct {
	User *authentication.User
	Data map[string]any
}

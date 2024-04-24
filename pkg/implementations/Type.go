package implementations

import (
	"smr/pkg/manager"
)

// Plugin contracts
type Implementation interface {
	Implementation(*manager.Manager, []byte) (Response, error)
}

type Response struct {
	HttpStatus       int
	Explanation      string
	ErrorExplanation string
	Error            bool
	Success          bool
}

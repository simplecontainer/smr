package template

import (
	"smr/pkg/config"
	"smr/pkg/runtime"
)

type Combinator struct {
	Configuration *config.Configuration
	Runtime       runtime.Runtime
}

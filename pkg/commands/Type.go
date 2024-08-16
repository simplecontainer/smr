package commands

import (
	"github.com/simplecontainer/smr/pkg/api"
)

type Command struct {
	name       string
	flag       string
	condition  func(*api.Api) bool
	functions  []func(*api.Api, []string)
	depends_on []func(*api.Api, []string)
}

package config

import (
	"github.com/simplecontainer/smr/pkg/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/manager"
	"net/http"
)

type Config struct {
	Started    bool
	Shared     *Shared
	Client     *http.Client
	Definition v1.ConfigurationDefinition
}

type Shared struct {
	Manager *manager.Manager
	Client  *client.Http
}

const KIND string = "configuration"

var invalidOperators []string = []string{
	"Run",
	"ListSupported",
}

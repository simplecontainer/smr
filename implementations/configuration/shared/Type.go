package shared

import (
	"github.com/simplecontainer/smr/pkg/manager"
	"net/http"
)

type Shared struct {
	Manager *manager.Manager
	Client  *http.Client
}

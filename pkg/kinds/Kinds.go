package kinds

import (
	"errors"
	"github.com/simplecontainer/smr/pkg/kinds/certkey"
	"github.com/simplecontainer/smr/pkg/kinds/config"
	"github.com/simplecontainer/smr/pkg/kinds/container"
	"github.com/simplecontainer/smr/pkg/kinds/containers"
	"github.com/simplecontainer/smr/pkg/kinds/gitops"
	"github.com/simplecontainer/smr/pkg/kinds/httpauth"
	"github.com/simplecontainer/smr/pkg/kinds/resource"
)

func New(kind string) (Operator, error) {
	switch kind {
	case "certkey":
		return certkey.New(), nil
	case "config":
		return config.New(), nil
	case "container":
		return container.New(), nil
	case "containers":
		return containers.New(), nil
	case "gitops":
		return gitops.New(), nil
	case "httpauth":
		return httpauth.New(), nil
	case "resource":
		return resource.New(), nil
	default:
		return nil, errors.New("kind does not exist")
	}
}

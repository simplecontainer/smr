package kinds

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/kinds/certkey"
	"github.com/simplecontainer/smr/pkg/kinds/config"
	"github.com/simplecontainer/smr/pkg/kinds/container"
	"github.com/simplecontainer/smr/pkg/kinds/containers"
	"github.com/simplecontainer/smr/pkg/kinds/custom"
	"github.com/simplecontainer/smr/pkg/kinds/gitops"
	"github.com/simplecontainer/smr/pkg/kinds/httpauth"
	"github.com/simplecontainer/smr/pkg/kinds/network"
	"github.com/simplecontainer/smr/pkg/kinds/resource"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
)

func New(kind string, mgr *manager.Manager) (contracts.Kind, error) {
	switch kind {
	case "custom":
		return custom.New(mgr), nil
	case "certkey":
		return certkey.New(mgr), nil
	case "configuration":
		return config.New(mgr), nil
	case "container":
		return container.New(mgr), nil
	case "containers":
		return containers.New(mgr), nil
	case "gitops":
		return gitops.New(mgr), nil
	case "httpauth":
		return httpauth.New(mgr), nil
	case "network":
		return network.New(mgr), nil
	case "resource":
		return resource.New(mgr), nil
	default:
		return nil, errors.New(fmt.Sprintf("%s kind does not exist", kind))
	}
}

func BuildRegistry(mgr *manager.Manager) map[string]contracts.Kind {
	var kindsRegistry = make(map[string]contracts.Kind, 0)
	var err error

	for kind, _ := range mgr.Kinds.Relations {
		kindsRegistry[kind], err = New(kind, mgr)

		if err != nil {
			panic(err)
		}

		err = kindsRegistry[kind].Start()

		if err != nil {
			panic(err)
		}

		logger.Log.Info(fmt.Sprintf("registered and started kind: %s", kind))
	}

	return kindsRegistry
}

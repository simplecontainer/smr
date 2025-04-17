package kinds

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts/ikinds"
	"github.com/simplecontainer/smr/pkg/kinds/certkey"
	"github.com/simplecontainer/smr/pkg/kinds/config"
	"github.com/simplecontainer/smr/pkg/kinds/containers"
	"github.com/simplecontainer/smr/pkg/kinds/custom"
	"github.com/simplecontainer/smr/pkg/kinds/gitops"
	"github.com/simplecontainer/smr/pkg/kinds/httpauth"
	"github.com/simplecontainer/smr/pkg/kinds/network"
	"github.com/simplecontainer/smr/pkg/kinds/node"
	"github.com/simplecontainer/smr/pkg/kinds/resource"
	"github.com/simplecontainer/smr/pkg/kinds/secret"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
)

func New(kind string, mgr *manager.Manager) (ikinds.Kind, error) {
	switch kind {
	case "node":
		return node.New(mgr), nil
	case "custom":
		return custom.New(mgr), nil
	case "certkey":
		return certkey.New(mgr), nil
	case "configuration":
		return config.New(mgr), nil
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
	case "secret":
		return secret.New(mgr), nil
	default:
		return nil, errors.New(fmt.Sprintf("%s kind does not exist", kind))
	}
}

func BuildRegistry(mgr *manager.Manager) map[string]ikinds.Kind {
	var kindsRegistry = make(map[string]ikinds.Kind)
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

package definitions

import (
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/static"
)

func ClusterNetwork(subnetCIDR string) *v1.NetworkDefinition {
	definition := &v1.NetworkDefinition{
		Kind:   static.KIND_NETWORK,
		Prefix: static.SMR_PREFIX,
		Meta: commonv1.Meta{
			Group: "internal",
			Name:  "cluster",
		},
		Spec: v1.NetworkSpec{
			Driver:          "bridge",
			IPV4AddressPool: subnetCIDR,
		},
		State: commonv1.NewState(),
	}

	definition.State.AddOpt("scope", "local")
	return definition
}

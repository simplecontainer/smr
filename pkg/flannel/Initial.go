package flannel

import (
	"context"
	"fmt"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/static"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func Setup(ctx context.Context, client *clientv3.Client, network string, backend string) error {
	_, err := client.Put(ctx, "/coreos.com/network/config", fmt.Sprintf("{\"Network\": \"%s\", \"Backend\": {\"Type\": \"%s\"}}", network, backend))
	return err
}

func Definition(subnetCIDR string) *v1.NetworkDefinition {
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

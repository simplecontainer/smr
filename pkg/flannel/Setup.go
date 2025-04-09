package flannel

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func Setup(ctx context.Context, client *clientv3.Client, network string, backend string) error {
	_, err := client.Put(ctx, "/coreos.com/network/config", fmt.Sprintf("{\"Network\": \"%s\", \"Backend\": {\"Type\": \"%s\"}}", network, backend))
	return err
}

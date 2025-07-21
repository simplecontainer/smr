package flannel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/simplecontainer/smr/internal/definitions"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/static"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

const (
	emptyIPv6Network = "::/0"
	etcdDialTimeout  = 5 * time.Second
	etcdWatchPrefix  = "/coreos.com/network/subnets"
)

const (
	ipv4 = 1 << iota
	ipv6
)

func Run(ctx context.Context, cancel context.CancelFunc, c *client.Client, config *configuration.Configuration) error {
	logger.Log.Info("starting flannel with backend", zap.String("backend", config.Flannel.Backend))

	f, err := initializeFlannel(config)
	if err != nil {
		return errors.Wrap(err, "failed to initialize flannel")
	}

	go func() {
		err := flannel(ctx, f, f.InterfaceSpecified, f.IPv6Masq, f.NetMode)
		if err != nil {
			logger.Log.Error("flannel exited", zap.Error(err))
		}
		cancel()
	}()

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{fmt.Sprintf("localhost:%s", config.Ports.Etcd)},
		DialTimeout: etcdDialTimeout,
	})
	if err != nil {
		return errors.Wrap(err, "failed to connect to etcd")
	}
	defer cli.Close()

	return watchSubnets(ctx, cli, c, f)
}

func initializeFlannel(config *configuration.Configuration) (*Flannel, error) {
	f := New(subnetFile)

	if err := f.Clear(); err != nil {
		return nil, errors.Wrap(err, "failed to clear subnet file")
	}

	if err := f.SetBackend(config.Flannel.Backend); err != nil {
		return nil, errors.Wrap(err, "failed to set backend")
	}

	if err := f.EnableIPv4(config.Flannel.EnableIPv4); err != nil {
		return nil, errors.Wrap(err, "failed to enable IPv4")
	}

	if err := f.EnableIPv6(config.Flannel.EnableIPv6); err != nil {
		return nil, errors.Wrap(err, "failed to enable IPv6")
	}

	f.MaskIPv6(config.Flannel.IPv6Masq)

	if err := f.SetCIDR(config.Flannel.CIDR); err != nil {
		return nil, errors.Wrap(err, "failed to set CIDR")
	}

	if err := f.SetInterface(config.Flannel.InterfaceSpecified); err != nil {
		return nil, errors.Wrap(err, "failed to set interface")
	}

	netMode, err := findNetMode(f.CIDR)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check netMode for flannel")
	}

	f.NetMode = netMode

	return f, nil
}

func watchSubnets(ctx context.Context, cli *clientv3.Client, c *client.Client, f *Flannel) error {
	watcher := cli.Watch(ctx, etcdWatchPrefix, clientv3.WithPrefix())
	logger.Log.Info("client will wait for flannel to return subnet range")

	// Store processed events to prevent duplicates
	processedEvents := make(map[string]string)

	for {
		select {
		case watchResp, ok := <-watcher:
			if !ok {
				return errors.New("watcher channel closed unexpectedly")
			}

			for _, event := range watchResp.Events {
				if event.Type != mvccpb.PUT || !strings.Contains(string(event.Kv.Key), "subnet") || event.Kv.Lease == 0 {
					continue
				}

				// Process subnet event
				if err := processSubnetEvent(ctx, cli, c, f, event, processedEvents); err != nil {
					logger.Log.Error("failed to process subnet event", zap.Error(err), zap.String("key", string(event.Kv.Key)))
				}
			}

		case <-ctx.Done():
			return errors.New("context canceled - shutting down flannel watcher")
		}
	}
}

func processSubnetEvent(ctx context.Context, cli *clientv3.Client, c *client.Client, f *Flannel,
	event *clientv3.Event, processedEvents map[string]string) error {

	var subnet Subnet
	if err := json.Unmarshal(event.Kv.Value, &subnet); err != nil {
		return errors.Wrap(err, "failed to unmarshal subnet data")
	}

	logger.Log.Info("got subnet", zap.Any("subnet", subnet))

	if shouldCreateNetworkObject(f, subnet) {
		if err := createNetworkObject(c, event.Kv.Key); err != nil {
			return errors.Wrap(err, "failed to create network object")
		}
	}

	eventKey := string(event.Kv.Key)
	eventValue := string(event.Kv.Value)
	if processedEvents[eventKey] == eventValue {
		return nil
	}

	processedEvents[eventKey] = eventValue

	response := network.Send(
		c.Context.GetHTTPClient(),
		fmt.Sprintf("%s/api/v1/key/propose/%s", c.Context.APIURL, event.Kv.Key),
		http.MethodPost,
		event.Kv.Value,
	)

	if !response.Success {
		logger.Log.Error("flannel failed to inform members about subnet decision - abort startup")
		os.Exit(1)
	}

	go keepAliveSubnet(ctx, cli, event.Kv.Lease)

	return nil
}

func shouldCreateNetworkObject(f *Flannel, subnet Subnet) bool {
	switch {
	case f.NetMode == ipv4 && f.Interface.ExtAddr.String() == subnet.PublicIP:
		return true
	case f.NetMode == ipv6 && f.Interface.ExtV6Addr.String() == subnet.PublicIPv6:
		return true
	default:
		return false
	}
}

func createNetworkObject(c *client.Client, subnetKey []byte) error {
	split := strings.Split(string(subnetKey), "/")
	cidr := strings.Replace(split[len(split)-1], "-", "/", 1)

	networkDefinition, err := definitions.ClusterNetwork(cidr).ToJSON()
	if err != nil {
		return errors.Wrap(err, "failed to create network definition")
	}

	req, err := common.NewRequest(static.KIND_NETWORK)
	if err != nil {
		return errors.Wrap(err, "failed to create network request")
	}

	if err := req.Definition.FromJson(networkDefinition); err != nil {
		return errors.Wrap(err, "failed to parse network definition")
	}

	if err := req.ProposeApply(c.Context.GetHTTPClient(), c.Context.APIURL); err != nil {
		return errors.Wrap(err, "failed to apply network object")
	}

	logger.Log.Info("network object applied successfully")
	return nil
}

func keepAliveSubnet(ctx context.Context, cli *clientv3.Client, leaseID int64) {
	kach, err := cli.KeepAlive(ctx, clientv3.LeaseID(leaseID))
	if err != nil {
		logger.Log.Error("failed to start keepalive", zap.Error(err), zap.Int64("leaseID", leaseID))
		return
	}

	for {
		select {
		case resp, ok := <-kach:
			if !ok {
				logger.Log.Info("closed keep alive channel for lease", zap.Int64("leaseID", leaseID))
				return
			}
			logger.Log.Debug("keep alive response received", zap.String("response", resp.String()))

		case <-ctx.Done():
			logger.Log.Info("context canceled, stopping keepalive", zap.Int64("leaseID", leaseID))
			return
		}
	}
}

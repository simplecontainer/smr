package drain

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/contracts/icontrol"
	"github.com/simplecontainer/smr/pkg/control/generic"
	"github.com/simplecontainer/smr/pkg/control/registry"
	"github.com/simplecontainer/smr/pkg/events/events"
	nshared "github.com/simplecontainer/smr/pkg/kinds/node/shared"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/static"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.uber.org/zap"
	"time"
)

type Command struct {
	*generic.GenericCommand
}

func init() {
	registry.RegisterCommandType("drain", func() icontrol.Command {
		return &Command{
			GenericCommand: &generic.GenericCommand{},
		}
	})
}

func NewDrainCommand(options map[string]string) *Command {
	return &Command{
		GenericCommand: generic.NewCommand("drain", options),
	}
}

func (c *Command) Node(api iapi.Api, params map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), configuration.Timeout.CompleteDrainTimeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	api.GetCluster().Node.State.ModifyControl("draining", node.StatusInProgress)

	event, err := events.NewNodeEvent(events.EVENT_DRAIN_STARTED, api.GetCluster().Node)

	if err != nil {
		return err
	}

	logger.Log.Info("dispatched node event", zap.String("event", event.GetType()))

	events.Dispatch(event, api.GetKindsRegistry()[static.KIND_NODE].GetShared().(*nshared.Shared), api.GetCluster().Node.NodeID)

	for _, value := range api.GetKindsRegistry() {
		value.GetShared().Drain()

		timeout := time.After(configuration.Timeout.ResourceDrainTimeout)
		ticker := time.NewTicker(10 * time.Millisecond)

		for !value.GetShared().IsDrained() {
			select {
			case <-timeout:
				ticker.Stop()
				return errors.New("timed out waiting for drain to complete for the kind")
			case <-ticker.C:
			}
		}

		ticker.Stop()
	}

	for {
		select {
		case <-ctx.Done():
			event, err := events.NewNodeEvent(events.EVENT_DRAIN_FAILED, api.GetCluster().Node)

			if err != nil {
				return err
			}

			logger.Log.Info("dispatched node event", zap.String("event", event.GetType()))

			events.Dispatch(event, api.GetKindsRegistry()[static.KIND_NODE].GetShared().(*nshared.Shared), api.GetCluster().Node.NodeID)

			return errors.New("draining timeout exceeded, drain aborted - restart will occur")

		case <-ticker.C:
			if isAllWatchersDrained(api) {
				ticker.Stop()

				event, err := events.NewNodeEvent(events.EVENT_DRAIN_SUCCESS, api.GetCluster().Node)

				if err != nil {
					return err
				}

				logger.Log.Info("dispatched node event", zap.String("event", event.GetType()))

				format := event.ToFormat().ToString()
				api.GetReplication().Informer.AddCh(format)

				events.Dispatch(event, api.GetKindsRegistry()[static.KIND_NODE].GetShared().(*nshared.Shared), api.GetCluster().Node.NodeID)

				select {
				case <-api.GetReplication().Informer.GetCh(format):
					break
				case <-ctx.Done():
					return errors.New("timed out waiting for event acknowledgment")
				}

				var bytes []byte
				bytes, err = json.Marshal(c)

				if err != nil {
					return err
				}

				api.GetCluster().Node.ConfChange = raftpb.ConfChange{
					Type:    raftpb.ConfChangeRemoveNode,
					NodeID:  c.NodeID(),
					Context: bytes,
				}

				logger.Log.Info("node trigger snapshot before shutting down raft", zap.Uint64("node", c.NodeID()))
				err = api.GetCluster().RaftNode.ForceSnapshot()

				if err != nil {
					logger.Log.Error("failed to trigger snapshot", zap.Error(err))
				}

				api.GetCluster().NodeConf <- *api.GetCluster().Node
			}
			break

		case finalized := <-api.GetCluster().NodeFinalizer:
			var cmd generic.GenericCommand

			if err := json.Unmarshal(finalized.ConfChange.Context, &cmd); err != nil {
				logger.Log.Info("invalid finalizer context", zap.Error(err))
				continue
			}

			if cmd.Time() == c.Time() {
				logger.Log.Info("finalized node drain", zap.Uint64("node", finalized.NodeID))
				api.GetCluster().Node.State.ModifyControl("draining", node.StatusSuccess)

				return nil
			} else {
				return errors.New("timestamp mismatch in finalizer")
			}
		}
	}
}

func (c *Command) Agent(api iapi.Api, params map[string]string) error {
	return nil
}

func isAllWatchersDrained(api iapi.Api) bool {
	for _, value := range api.GetKindsRegistry() {
		if !value.GetShared().IsDrained() {
			return false
		}
	}

	return true
}

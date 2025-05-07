package drain

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/control/controls"
	"github.com/simplecontainer/smr/pkg/events/events"
	nshared "github.com/simplecontainer/smr/pkg/kinds/node/shared"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/static"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.uber.org/zap"
	"time"
)

type Command struct {
	*controls.GenericCommand
}

func NewDrainCommand(options map[string]string) *Command {
	return &Command{
		GenericCommand: controls.NewCommand("drain", options),
	}
}

func (c *Command) Node(mgr *manager.Manager, params map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 160*time.Second)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	mgr.Cluster.Node.State.ModifyControl("draining", node.StatusInProgress)

	for _, value := range mgr.KindsRegistry {
		value.GetShared().Drain()
	}

	for {
		select {
		case <-ctx.Done():
			return errors.New("draining timeout exceeded 160 seconds, drain aborted - manual intervention needed")

		case <-ticker.C:
			if isAllWatchersDrained(mgr) {
				ticker.Stop()

				event, err := events.NewNodeEvent(events.EVENT_DRAIN_STARTED, mgr.Cluster.Node)

				if err != nil {
					return err
				}

				logger.Log.Info("dispatched node event", zap.String("event", event.GetType()))

				format := event.ToFormat().ToString()
				mgr.Replication.Informer.AddCh(format)

				events.Dispatch(event, mgr.KindsRegistry[static.KIND_NODE].GetShared().(*nshared.Shared), mgr.Cluster.Node.NodeID)

				select {
				case <-mgr.Replication.Informer.GetCh(format):
					break
				case <-ctx.Done():
					return errors.New("timed out waiting for event acknowledgment")
				}

				var bytes []byte
				bytes, err = json.Marshal(c)

				if err != nil {
					return err
				}

				mgr.Cluster.Node.ConfChange = raftpb.ConfChange{
					Type:    raftpb.ConfChangeRemoveNode,
					NodeID:  c.NodeID(),
					Context: bytes,
				}

				mgr.Cluster.NodeConf <- *mgr.Cluster.Node
			}
			break

		case finalized := <-mgr.Cluster.NodeFinalizer:
			var cmd controls.Command

			if err := json.Unmarshal(finalized.ConfChange.Context, &cmd); err != nil {
				logger.Log.Info("invalid finalizer context", zap.Error(err))
				continue
			}

			if cmd.Time() == c.Time() {
				logger.Log.Info("finalizing node", zap.Uint64("node", finalized.NodeID))
				mgr.Cluster.Node.State.ModifyControl("draining", node.StatusSuccess)

				return nil
			} else {
				return errors.New("timestamp mismatch in finalizer")
			}
		}
	}
}

func (c *Command) Agent(cli *client.Client, params map[string]string) error {
	return nil
}

func isAllWatchersDrained(mgr *manager.Manager) bool {
	for _, value := range mgr.KindsRegistry {
		if !value.GetShared().IsDrained() {
			return false
		}
	}

	return true
}

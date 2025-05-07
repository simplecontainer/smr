package distributed

import (
	"fmt"
	"github.com/simplecontainer/smr/internal/helpers"
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/acks"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/contracts/iformat"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/node"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/smaps"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"strings"
)

func New(client *clients.Client, user *authentication.User, nodeName string, node *node.Node) *Replication {
	return &Replication{
		Client:     client,
		Node:       node,
		User:       user,
		DataC:      make(chan KV.KV),
		Informer:   NewInformer(),
		Replicated: smaps.New(),
	}
}

func (replication *Replication) ListenData(agent string) {
	for {
		select {
		case data, ok := <-replication.DataC:
			if ok {
				format := f.NewFromString(data.Key)

				if !format.IsValid() {
					logger.Log.Error("invalid format distributed", zap.String("format", data.Key))
				} else {
					switch format.GetCategory() {
					case static.CATEGORY_PLAIN:
						replication.HandlePlain(format, data)
						break
					case static.CATEGORY_STATE:
						replication.HandlePlain(format, data)
						break
					case static.CATEGORY_KIND:
						replication.HandleObject(format, data)
						break
					case static.CATEGORY_DNS:
						replication.DnsUpdatesC <- data
						break
					case static.CATEGORY_EVENT:
						replication.EventsC <- data
						break
					default:
						replication.HandleOutside(data)
						break
					}
				}
				break
			}
		}

	}
}

func (replication *Replication) HandleObject(format iformat.Format, data KV.KV) {
	defer acks.ACKS.Ack(format.GetUUID())

	request, _ := common.NewRequest(format.GetKind())
	err := request.Definition.FromJson(data.Val)

	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	if request.Definition.GetState() != nil {
		if request.Definition.GetState().GetOpt("scope").Value == "local" {
			if data.Node != replication.Node.NodeID {
				logger.Log.Info("locally scoped object only", zap.Uint64("node", data.Node))
				return
			}
		}
	}

	action := request.Definition.GetState().GetOpt("action").Value
	request.Definition.GetState().ClearOpt("action")

	switch action {
	case static.STATE_KIND:
		helpers.LogIfError(request.AttemptState(replication.Client.Http, replication.Client.API))
		break
	case static.REMOVE_KIND:
		helpers.LogIfError(request.AttemptRemove(replication.Client.Http, replication.Client.API))
		break
	default:
		helpers.LogIfError(request.AttemptApply(replication.Client.Http, replication.Client.API))
		break
	}
}

func (replication *Replication) HandlePlain(format iformat.Format, data KV.KV) {
	defer acks.ACKS.Ack(format.GetUUID())
	obj := objects.New(replication.Client, replication.User)

	if data.Val == nil {
		_, err := obj.RemoveLocal(format)

		if err != nil {
			logger.Log.Error(err.Error())
		}
	} else {
		err := obj.AddLocal(format, data.Val)

		if err != nil {
			logger.Log.Error(err.Error())
		}
	}
}

func (replication *Replication) HandleDns(format iformat.Format, data KV.KV) {
	defer acks.ACKS.Ack(format.GetUUID())

	obj := objects.New(replication.Client, replication.User)

	if data.Val == nil {
		_, err := obj.RemoveLocal(format)

		if err != nil {
			logger.Log.Error(err.Error())
		}
	} else {
		err := obj.AddLocal(format, data.Val)

		if err != nil {
			logger.Log.Error(err.Error())
		}
	}
}

// HandleEtcd handles the case when data is entered into etcd via other means than simplecontainer - flannel only
func (replication *Replication) HandleOutside(data KV.KV) {
	obj := objects.New(replication.Client, replication.User)

	format := f.NewUnformated(strings.TrimPrefix(data.Key, "/"))
	key := fmt.Sprintf("/%s", format.ToString())

	replication.Replicated.Map.Store(key, 1)

	if !data.IsLocal() {
		if data.Val == nil {
			_, err := obj.RemoveLocalKey(key)

			if err != nil {
				logger.Log.Error(err.Error())
			}
		} else {
			err := obj.AddLocalKey(key, data.Val)

			if err != nil {
				logger.Log.Error(err.Error())
			}
		}
	}
}

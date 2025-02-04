package distributed

import (
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/acks"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/helpers"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/smaps"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
)

func New(client *client.Client, user *authentication.User, nodeName string) *Replication {
	return &Replication{
		Client:     client,
		NodeName:   nodeName,
		User:       user,
		DataC:      make(chan KV.KV),
		Replicated: smaps.New(),
	}
}

func (replication *Replication) ListenData(agent string) {
	for {
		select {
		case data, ok := <-replication.DataC:
			if ok {
				go func() {
					format := f.NewFromString(data.Key)

					if !format.IsValid() {
						logger.Log.Error("invalid format distributed", zap.String("format", data.Key))
					} else {
						switch format.GetCategory() {
						case static.CATEGORY_PLAIN:
							replication.HandlePlain(format, data)
							break
						case static.CATEGORY_SECRET:
							replication.HandleSecret(format, data)
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
				}()
				break
			}
		}

	}
}

func (replication *Replication) HandleObject(format contracts.Format, data KV.KV) {
	acks.ACKS.Ack(format.GetUUID())

	request, _ := common.NewRequest(format.GetKind())
	err := request.Definition.FromJson(data.Val)

	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	request.Definition.GetRuntime().SetNode(data.Node)

	switch request.Definition.GetState().GetOpt("action").Value {
	case static.REMOVE_KIND:
		helpers.LogIfError(request.AttemptRemove(replication.Client.Http, replication.Client.API))
		break
	default:
		helpers.LogIfError(request.AttemptApply(replication.Client.Http, replication.Client.API))
		break
	}
}

func (replication *Replication) HandlePlain(format contracts.Format, data KV.KV) {
	acks.ACKS.Ack(format.GetUUID())

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

func (replication *Replication) HandleSecret(format contracts.Format, data KV.KV) {
	acks.ACKS.Ack(format.GetUUID())

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

func (replication *Replication) HandleDns(format contracts.Format, data KV.KV) {
	acks.ACKS.Ack(format.GetUUID())

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

	replication.Replicated.Map.Store(data.Key, 1)

	if !data.IsLocal() {
		if data.Val == nil {
			_, err := obj.RemoveLocalKey(data.Key)

			if err != nil {
				logger.Log.Error(err.Error())
			}
		} else {
			err := obj.AddLocalKey(data.Key, data.Val)

			if err != nil {
				logger.Log.Error(err.Error())
			}
		}
	}
}

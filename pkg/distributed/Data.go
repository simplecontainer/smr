package distributed

import (
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/acks"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/secrets"
	"github.com/simplecontainer/smr/pkg/smaps"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"net/http"
	"strings"
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
					switch data.Category {
					case static.CATEGORY_PLAIN:
						replication.HandlePlain(data)
						break
					case static.CATEGORY_SECRET:
						replication.HandleSecret(data)
						break
					case static.CATEGORY_ETCD:
						replication.HandleEtcd(data)
						break
					case static.CATEGORY_OBJECT:
						replication.HandleObject(data)
						break
					case static.CATEGORY_OBJECT_DELETE:
						replication.HandleObjectDelete(data)
						break
					case static.CATEGORY_DNS:
						replication.DnsUpdatesC <- data
						break
					case static.CATEGORY_EVENT:
						replication.EventsC <- data
						break
					case static.CATEGORY_INVALID:
						replication.HandleInvalid(data)
						break
					default:
						break
					}

				}()
				break
			}
		}

	}
}

func (replication *Replication) HandleObject(data KV.KV) {
	format := f.NewFromString(data.Key)
	acks.ACKS.Ack(format.GetUUID())

	request, _ := common.NewRequest(format.Kind)
	request.Definition.FromJson(data.Val)
	request.Definition.GetRuntime().SetNode(data.Node)

	bytes, err := request.Definition.ToJsonWithKind()

	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	if data.Val == nil {
		response := network.Send(replication.Client.Http, fmt.Sprintf("https://localhost:1443/api/v1/delete"), http.MethodPost, bytes)

		if response != nil {
			if !response.Success {
				if !strings.HasSuffix(response.ErrorExplanation, "object is same on the server") {
					logger.Log.Error(errors.New(response.ErrorExplanation).Error())
				}
			}
		}
	} else {
		response := network.Send(replication.Client.Http, fmt.Sprintf("https://localhost:1443/api/v1/apply"), http.MethodPost, bytes)

		if response != nil {
			if !response.Success {
				if !strings.HasSuffix(response.ErrorExplanation, "object is same on the server") {
					logger.Log.Error(errors.New(response.ErrorExplanation).Error())
				}
			}
		}
	}
}

func (replication *Replication) HandleObjectDelete(data KV.KV) {
	format := f.NewFromString(data.Key)
	acks.ACKS.Ack(format.GetUUID())

	request, _ := common.NewRequest(format.Kind)
	request.Definition.FromJson(data.Val)
	request.Definition.GetRuntime().SetNode(data.Node)

	bytes, err := request.Definition.ToJsonWithKind()

	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	response := network.Send(replication.Client.Http, fmt.Sprintf("https://localhost:1443/api/v1/delete"), http.MethodDelete, bytes)

	if response != nil {
		if !response.Success {
			if !strings.HasSuffix(response.ErrorExplanation, "object is same on the server") {
				logger.Log.Error(errors.New(response.ErrorExplanation).Error())
			}
		}
	}
}

func (replication *Replication) HandlePlain(data KV.KV) {
	format := f.NewUnformated(data.Key, static.CATEGORY_PLAIN_STRING)
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
func (replication *Replication) HandleEtcd(data KV.KV) {
	format := f.NewUnformated(data.Key, static.CATEGORY_ETCD_STRING)
	acks.ACKS.Ack(format.GetUUID())

	obj := objects.New(replication.Client, replication.User)

	replication.Replicated.Map.Store(format.ToString(), 1)

	if !data.IsLocal() {
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
}

func (replication *Replication) HandleSecret(data KV.KV) {
	format := f.NewUnformated(data.Key, static.CATEGORY_SECRET_STRING)
	acks.ACKS.Ack(format.GetUUID())

	obj := secrets.New(replication.Client, replication.User)

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

func (replication *Replication) HandleDns(data KV.KV) {
	format := f.NewUnformated(data.Key, static.CATEGORY_DNS_STRING)
	acks.ACKS.Ack(format.GetUUID())

	obj := secrets.New(replication.Client, replication.User)

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

func (replication *Replication) HandleInvalid(data KV.KV) {
	logger.Log.Error("invalid replication category", zap.String("key", data.Key), zap.String("value", string(data.Val)), zap.Int("Category", data.Category))
}

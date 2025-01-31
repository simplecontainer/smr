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
					format := f.NewFromString(data.Key)

					switch format.Category {
					case static.CATEGORY_PLAIN:
						replication.HandlePlain(data)
						break
					case static.CATEGORY_SECRET:
						replication.HandleSecret(data)
						break
					case static.CATEGORY_STATE:
						replication.HandlePlain(data)
						break
					case static.CATEGORY_KIND:
						replication.HandleObject(data)
						break
					default:
						replication.HandleOutside(data)
						break
					case static.CATEGORY_DNS:
						replication.DnsUpdatesC <- data
						break
					case static.CATEGORY_EVENT:
						replication.EventsC <- data
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

func (replication *Replication) HandlePlain(data KV.KV) {
	format := f.NewFromString(data.Key)
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

func (replication *Replication) HandleSecret(data KV.KV) {
	format := f.NewFromString(data.Key)
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
	format := f.NewFromString(data.Key)
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

// HandleEtcd handles the case when data is entered into etcd via other means than simplecontainer - flannel only
func (replication *Replication) HandleOutside(data KV.KV) {
	format := f.NewUnformated(data.Key)
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

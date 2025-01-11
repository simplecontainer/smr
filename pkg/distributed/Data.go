package distributed

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
)

func New(client *client.Client, user *authentication.User) *Replication {
	return &Replication{
		Client: client,
		User:   user,
		DataC:  make(chan KV),
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

func (replication *Replication) HandleObject(data KV) {
	format := f.NewFromString(data.Key)
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

func (replication *Replication) HandlePlain(data KV) {
	format := f.NewUnformated(data.Key, static.CATEGORY_PLAIN_STRING)
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

func (replication *Replication) HandleEtcd(data KV) {
	format := f.NewUnformated(data.Key, static.CATEGORY_ETCD_STRING)
	obj := objects.New(replication.Client, replication.User)

	if data.Val == nil {
		_, err := obj.RemoveLocal(format)

		if err != nil {
			logger.Log.Error(err.Error())
		}

		err = EtcDelete(data.Key)

		if err != nil {
			logger.Log.Error(err.Error())
		}
	} else {
		err := obj.AddLocal(format, data.Val)

		if err != nil {
			logger.Log.Error(err.Error())
		}

		err = EtcdPut(data.Key, string(data.Val))

		if err != nil {
			logger.Log.Error(err.Error())
		}
	}
}

func (replication *Replication) HandleSecret(data KV) {
	format := f.NewUnformated(data.Key, static.CATEGORY_SECRET_STRING)
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

func (replication *Replication) HandleInvalid(data KV) {
	logger.Log.Error("invalid replication category", zap.String("key", data.Key), zap.String("value", string(data.Val)), zap.Int("Category", data.Category))
}

package distributed

import (
	"bytes"
	"context"
	"errors"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func (replication *Replication) ListenEtcd(agent string) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return nil
	}

	ctx, _ := context.WithCancel(context.Background())
	watcher := cli.Watch(ctx, "/coreos.com", clientv3.WithPrefix())

	for {
		select {
		case watchResp, ok := <-watcher:
			if ok {
				for _, event := range watchResp.Events {
					switch event.Type {
					case mvccpb.PUT:
						obj := objects.New(replication.Client, replication.User)
						format := f.NewUnformated(string(event.Kv.Key), static.CATEGORY_ETCD_STRING)

						_, err = obj.Propose(format, event.Kv.Value)

						if err != nil {
							logger.Log.Error(err.Error())
						}
						break
					case mvccpb.DELETE:
						obj := objects.New(replication.Client, replication.User)
						format := f.NewUnformated(string(event.Kv.Key), static.CATEGORY_ETCD_STRING)

						_, err = obj.Propose(format, nil)

						if err != nil {
							logger.Log.Error(err.Error())
						}
						break
					}
				}
			}
		case <-ctx.Done():
			logger.Log.Error(errors.New("closed watcher channel should not block").Error())
		}
	}
}

func EtcdPut(key string, value string) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return err
	}

	ctx, _ := context.WithCancel(context.Background())

	etcdValue, err := cli.Get(ctx, key)

	if err != nil {
		logger.Log.Error(err.Error())
	}

	// Put only if applied value is different from the previous value to avoid trigger replication flow again
	if len(etcdValue.Kvs) == 0 || !bytes.Equal(etcdValue.Kvs[len(etcdValue.Kvs)-1].Value, []byte(value)) {
		_, err = cli.Put(ctx, key, value)

		if err != nil {
			return err
		}
	}

	return nil
}

func EtcDelete(key string) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return err
	}

	ctx, _ := context.WithCancel(context.Background())

	_, err = cli.Delete(ctx, key)

	if err != nil {
		return err
	}

	return nil
}

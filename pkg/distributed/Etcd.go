package distributed

import (
	"context"
	"errors"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func (replication *Replication) ListenEtcd(agent string) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	ctx, _ := context.WithCancel(context.Background())
	watcher := cli.Watch(ctx, "/coreos.com", clientv3.WithPrefix())

	for {
		select {
		case watchResp, ok := <-watcher:
			if ok {
				for _, event := range watchResp.Events {
					_, ok := replication.Replicated.Map.Load(string(event.Kv.Key))

					if ok {
						replication.Replicated.Map.Delete(string(event.Kv.Key))
					} else {
						switch event.Type {
						case mvccpb.PUT:
							obj := objects.New(replication.Client, replication.User)
							format := f.NewUnformated(string(event.Kv.Key))

							err = obj.Propose(format, event.Kv.Value)

							if err != nil {
								logger.Log.Error(err.Error())
							}
							break
						case mvccpb.DELETE:
							obj := objects.New(replication.Client, replication.User)
							format := f.NewUnformated(string(event.Kv.Key))

							err = obj.Propose(format, nil)

							if err != nil {
								logger.Log.Error(err.Error())
							}
							break
						}
					}
				}
			}
		case <-ctx.Done():
			logger.Log.Error(errors.New("closed watcher channel should not block").Error())
		}
	}
}

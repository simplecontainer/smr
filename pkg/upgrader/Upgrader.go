package upgrader

import (
	"context"
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func New(image string, tag string) *Upgrade {
	return &Upgrade{
		Image: image,
		Tag:   tag,
	}
}

func (u *Upgrade) Apply(ctx context.Context, client *clientv3.Client) error {
	bytes, err := u.ToJson()

	if err != nil {
		return err
	}

	_, err = client.Put(ctx, "/smr/upgrade", string(bytes))
	return err
}

func (u *Upgrade) ToJson() ([]byte, error) {
	return json.Marshal(u)
}

package icontrol

import (
	"context"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts/iapi"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type Command interface {
	Name() string
	Time() time.Time
	Node(iapi.Api, map[string]string) error
	Agent(iapi.Api, map[string]string) error
	Data() map[string]string
	NodeID() uint64
	SetNodeID(uint64)
}

type Batch interface {
	GetNodeID() uint64
	SetNodeID(uint64)
	Put(ctx context.Context, client *clientv3.Client) error
	Apply(ctx context.Context, cli *client.Client) (*iresponse.Response, error)
	AddCommand(cmd Command)
	GetCommands() []Command
	GetCommand(name string) (Command, error)
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error
}

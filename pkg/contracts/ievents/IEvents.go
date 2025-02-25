package ievents

import "github.com/simplecontainer/smr/pkg/raft"

type Event interface {
	GetType() string
	GetTarget() string
	GetKind() string
	GetGroup() string
	GetName() string
	GetData() []byte
	GetNetworkId() string
	GetContainerId() string
	ToJson() ([]byte, error)
	Propose(proposeC *raft.KVStore, node uint64) error
	IsManaged() bool
}

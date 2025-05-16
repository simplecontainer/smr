package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts/icontrol"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"github.com/simplecontainer/smr/pkg/control/registry"
	"github.com/simplecontainer/smr/pkg/network"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net/http"
	"time"
)

func NewCommandBatch() icontrol.Batch {
	return &CommandBatch{
		ID:        uuid.UUID{},
		Timestamp: time.Time{},
		Commands:  make([]icontrol.Command, 0),
		RawCmds:   make([]json.RawMessage, 0),
		NodeID:    0,
	}
}

func (b *CommandBatch) GetNodeID() uint64 {
	return b.NodeID
}

func (b *CommandBatch) SetNodeID(nodeID uint64) {
	b.NodeID = nodeID
}

func (b *CommandBatch) Put(ctx context.Context, client *clientv3.Client) error {
	bytes, err := json.Marshal(b)

	if err != nil {
		return err
	}

	_, err = client.Put(ctx, fmt.Sprintf("/smr/control/%s", b.ID), string(bytes))

	return err
}

func (b *CommandBatch) Apply(ctx context.Context, cli *client.Client) (*iresponse.Response, error) {
	bytes, err := json.Marshal(b)

	if err != nil {
		return nil, err
	}

	response := network.Send(cli.Context.GetClient(), fmt.Sprintf("%s/api/v1/cluster/control", cli.Context.APIURL), http.MethodPost, bytes)

	object := json.RawMessage{}

	err = json.Unmarshal(response.Data, &object)

	if err != nil {
		return response, err
	}

	return response, nil
}

func (b *CommandBatch) AddCommand(cmd icontrol.Command) {
	cmd.SetNodeID(b.NodeID)
	b.Commands = append(b.Commands, cmd)
}

func (b *CommandBatch) GetCommands() []icontrol.Command {
	return b.Commands
}

func (b *CommandBatch) GetCommand(name string) (icontrol.Command, error) {
	for _, cmd := range b.Commands {
		if cmd.Name() == name {
			return cmd, nil
		}
	}

	return nil, errors.New("control commands are empty")
}

func (b *CommandBatch) MarshalJSON() ([]byte, error) {
	b.RawCmds = make([]json.RawMessage, len(b.Commands))
	for i, cmd := range b.Commands {
		data, err := json.Marshal(cmd)
		if err != nil {
			return nil, err
		}

		b.RawCmds[i] = data
	}

	type Alias CommandBatch
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(b),
	})
}

func (b *CommandBatch) UnmarshalJSON(data []byte) error {
	aux := &struct {
		ID        uuid.UUID         `json:"id"`
		Timestamp time.Time         `json:"timestamp"`
		RawCmds   []json.RawMessage `json:"commands"`
		NodeID    uint64            `json:"nodeID"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	b.Timestamp = aux.Timestamp
	b.RawCmds = aux.RawCmds
	b.NodeID = aux.NodeID
	b.ID = aux.ID

	b.Commands = make([]icontrol.Command, len(b.RawCmds))
	for i, raw := range b.RawCmds {
		cmd, err := registry.UnmarshalCommand(raw)
		if err != nil {
			return err
		}

		cmd.SetNodeID(b.NodeID)
		b.Commands[i] = cmd
	}

	return nil
}

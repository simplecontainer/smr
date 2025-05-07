package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/simplecontainer/smr/pkg/control/controls"
	"github.com/simplecontainer/smr/pkg/control/controls/drain"
	"github.com/simplecontainer/smr/pkg/control/controls/start"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func NewCommandBatch() *CommandBatch {
	return &CommandBatch{
		ID:        uuid.New(),
		Timestamp: time.Now(),
		Commands:  make([]controls.Command, 0),
		RawCmds:   make([]json.RawMessage, 0),
		NodeID:    0,
	}
}

func (b *CommandBatch) Apply(ctx context.Context, client *clientv3.Client) error {
	bytes, err := json.Marshal(b)

	if err != nil {
		return err
	}

	_, err = client.Put(ctx, fmt.Sprintf("/smr/control/%s", b.ID), string(bytes))

	return err
}

func (b *CommandBatch) AddCommand(cmd controls.Command) {
	cmd.SetNodeID(b.NodeID)
	b.Commands = append(b.Commands, cmd)
}

func (b *CommandBatch) GetCommands() []controls.Command {
	return b.Commands
}

func (b *CommandBatch) GetCommand(name string) (controls.Command, error) {
	for _, cmd := range b.Commands {
		if cmd.Name() == name {
			return cmd, nil
		}
	}

	return nil, errors.New("control command not found")
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
		Commands  []json.RawMessage `json:"commands"`
		NodeID    uint64            `json:"nodeID"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	b.Timestamp = aux.Timestamp
	b.RawCmds = aux.Commands
	b.NodeID = aux.NodeID
	b.ID = aux.ID

	b.Commands = make([]controls.Command, len(b.RawCmds))
	for i, raw := range b.RawCmds {
		var peek struct {
			Name string `json:"name"`
		}

		if err := json.Unmarshal(raw, &peek); err != nil {
			return err
		}

		var cmd controls.Command

		switch peek.Name {
		case "drain":
			cmd = &drain.Command{
				GenericCommand: &controls.GenericCommand{},
			}
		case "start":
			cmd = &start.Command{
				GenericCommand: &controls.GenericCommand{},
			}
		default:
			return fmt.Errorf("unknown command type: %s", peek.Name)
		}

		if err := json.Unmarshal(raw, cmd); err != nil {
			return err
		}

		cmd.SetNodeID(b.NodeID)
		b.Commands[i] = cmd
	}

	return nil
}

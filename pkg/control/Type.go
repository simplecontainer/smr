package control

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/simplecontainer/smr/pkg/contracts/icontrol"
	"time"
)

type CommandBatch struct {
	ID        uuid.UUID          `json:"id"`
	Timestamp time.Time          `json:"timestamp"`
	Commands  []icontrol.Command `json:"-"`
	RawCmds   []json.RawMessage  `json:"commands"`
	NodeID    uint64             `json:"nodeID"`
}

package control

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/simplecontainer/smr/pkg/control/controls"
	"time"
)

type CommandBatch struct {
	ID        uuid.UUID
	Timestamp time.Time          `json:"timestamp"`
	Commands  []controls.Command `json:"-"`
	RawCmds   []json.RawMessage  `json:"commands"`
	NodeID    uint64             `json:"nodeID"`
}

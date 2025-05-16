package generic

import (
	"time"
)

type GenericCommand struct {
	nodeID    uint64
	name      string
	data      map[string]string
	timestamp time.Time
}

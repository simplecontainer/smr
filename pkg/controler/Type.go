package controler

import "time"

type Control struct {
	Drain     *Drain    `validate:"omitempty,dive" json:"drain,omitempty"`
	Upgrade   *Upgrade  `validate:"omitempty,dive" json:"upgrade,omitempty"`
	Start     *Start    `validate:"omitempty,dive" json:"start,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

type Drain struct {
	NodeID uint64 `validate:"required" json:"node_id"`
}

type Upgrade struct {
	Image string `validate:"required" json:"image"`
	Tag   string `validate:"required" json:"tag"`
}

type Start struct {
	NodeRaftAPI string `json:"node_raft_api" validate:"required"`
	Overlay     string `json:"overlay" validate:"required"`
	Backend     string `json:"backend" validate:"required"`
}

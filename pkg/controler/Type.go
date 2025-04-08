package controler

type Control struct {
	Drain   *Drain   `validate:"omitempty,dive" json:"drain,omitempty"`
	Upgrade *Upgrade `validate:"omitempty,dive" json:"upgrade,omitempty"`
	Start   *Start   `validate:"omitempty,dive" json:"start,omitempty"`
}

type Drain struct {
	NodeID uint64 `validate:"required" json:"node_id"`
}

type Upgrade struct {
	Image string `validate:"required" json:"image"`
	Tag   string `validate:"required" json:"tag"`
}

type Start struct {
	NodeAPI string `validate:"required" json:"node_api"`
	Overlay string `validate:"required" json:"overlay"`
	Backend string `validate:"required" json:"backend"`
}

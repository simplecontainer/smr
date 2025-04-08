package controler

type Control struct {
	Drain   *Drain   `validate:"required" json:"drain,omitempty"`
	Upgrade *Upgrade `validate:"required" json:"upgrade,omitempty"`
}

type Drain struct {
	NodeID uint64 `validate:"required" json:"node_id"`
}

type Upgrade struct {
	Image string `validate:"required" json:"image"`
	Tag   string `validate:"required" json:"tag"`
}

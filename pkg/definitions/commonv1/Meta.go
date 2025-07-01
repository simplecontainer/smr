package commonv1

type Meta struct {
	Group   string            `json:"group" validate:"required"`
	Name    string            `json:"name" validate:"required"`
	Labels  map[string]string `json:"labels,omitempty"`
	Runtime *Runtime          `json:"runtime,omitempty"`
}

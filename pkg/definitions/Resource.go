package definitions

type Resource struct {
	Meta ResourceMeta `mapstructure:"meta"`
	Spec ResourceSpec `mapstructure:"spec"`
}

type ResourceMeta struct {
	Group      string `json:"group"`
	Identifier string `json:"identifier"`
}

type ResourceSpec struct {
	Data map[string]any `json:"data"`
}

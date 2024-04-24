package definitions

type Configuration struct {
	Meta ConfigurationMeta `mapstructure:"meta"`
	Spec ConfigurationSpec `mapstructure:"spec"`
}

type ConfigurationMeta struct {
	Group      string `json:"group"`
	Identifier string `json:"identifier"`
}

type ConfigurationSpec struct {
	Data map[string]string `json:"data"`
}

package definitions

type HttpAuth struct {
	Meta HttpAuthMeta `mapstructure:"meta"`
	Spec HttpAuthSpec `mapstructure:"spec"`
}

type HttpAuthMeta struct {
	Group      string `json:"group"`
	Identifier string `json:"identifier"`
}

type HttpAuthSpec struct {
	Username string
	Password string
}

package definitions

type Gitops struct {
	Meta GitopsMeta `mapstructure:"meta"`
	Spec GitopsSpec `mapstructure:"spec"`
}

type GitopsMeta struct {
	Group      string `json:"group"`
	Identifier string `json:"identifier"`
}

type GitopsSpec struct {
	RepoURL         string      `json:"repoURL"`
	Revision        string      `json:"revision"`
	DirectoryPath   string      `json:"directory"`
	PoolingInterval string      `json:"poolingInterval"`
	AutomaticSync   bool        `json:"automaticSync"`
	CertKeyRef      CertKeyRef  `json:"certKeyRef"`
	HttpAuthRef     HttpauthRef `json:"httpAuthRef"`
}

type CertKeyRef struct {
	Group      string
	Identifier string
}

type HttpauthRef struct {
	Group      string
	Identifier string
}

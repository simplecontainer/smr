package v1

import "encoding/json"

type Gitops struct {
	Meta GitopsMeta `json:"meta"`
	Spec GitopsSpec `json:"spec"`
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

func (gitops *Gitops) ToJsonString() (string, error) {
	bytes, err := json.Marshal(gitops)
	return string(bytes), err
}

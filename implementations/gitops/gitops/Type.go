package gitops

import (
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/simplecontainer/smr/implementations/gitops/status"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"time"
)

const HTTP_AUTH_TYPE string = "http"
const SSH_AUTH_TYPE string = "ssh"

type Gitops struct {
	RepoURL         string
	Revision        string
	DirectoryPath   string
	PoolingInterval string
	LastPoll        time.Time
	ForcePoll       bool
	AutomaticSync   bool
	ManualSync      bool
	Commit          *object.Commit
	Path            string
	Status          *status.Status
	Auth            *Auth
	AuthInternal    *AuthInternal `json:"-"`
	API             string
	Context         string
	Definition      *v1.GitopsDefinition
}

type Auth struct {
	AuthType    *AuthType
	CertKeyRef  v1.GitopsCertKeyRef
	HttpAuthRef v1.GitopsHttpauthRef
}

type AuthType struct {
	AuthType string
}

type AuthInternal struct {
	CertKey  *CertKey  `json:"-"`
	HttpAuth *HttpAuth `json:"-"`
}

type HttpAuth struct {
	Username   string
	Password   string
	Definition v1.HttpAuthDefinition
}

type CertKey struct {
	Certificate        string `json:"certificate"`
	PublicKey          string `json:"publicKey"`
	PrivateKey         string `json:"privateKey"`
	PrivateKeyPassword string `json:"privateKeyPassword"`
	KeyStore           string `json:"keystore"`
	KeyStorePassword   string `json:"keyStorePassword"`
	CertStore          string `json:"certstore"`
	CertStorePassword  string `json:"certstorePassword"`
	Definition         v1.CertKeyDefinition
}

type Event struct {
	Event string
}

package client

import (
	"bytes"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/version"
	"net/http"
)

type Client struct {
	Config  *configuration.Configuration
	Context *Context
	Version *version.VersionClient
}

type Context struct {
	Directory     string
	ApiURL        string
	Name          string
	CertBundle    string
	PKCS12        string
	Ca            *bytes.Buffer `json:"-"`
	Cert          *bytes.Buffer `json:"-"`
	PrivateKey    *bytes.Buffer `json:"-"`
	Client        *http.Client  `json:"-"`
	ActiveContext string        `json:"-"`
}

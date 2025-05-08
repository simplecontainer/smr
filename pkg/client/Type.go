package client

import (
	"bytes"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/version"
	"net/http"
	"time"
)

type Client struct {
	Config  *configuration.Configuration
	Group   string
	Context *ClientContext
	Version *version.VersionClient
}

type Config struct {
	RootDir     string
	APITimeout  time.Duration
	MaxRetries  int
	RetryDelay  time.Duration
	UseInsecure bool
}

type Credentials struct {
	PrivateKey *bytes.Buffer `json:"-"`
	Cert       *bytes.Buffer `json:"-"`
	Ca         *bytes.Buffer `json:"-"`
	CertBundle string        `json:"cert_bundle"`
}

type ClientContext struct {
	Name       string `json:"name"`
	APIURL     string `json:"api_url"`
	Directory  string `json:"-"`
	ActivePath string `json:"active_path,omitempty"`

	Credentials *Credentials `json:"credentials"`
	client      *http.Client `json:"-"`
	config      *Config      `json:"-"`
}

type Manager struct {
	config *Config
}

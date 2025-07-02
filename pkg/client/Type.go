package client

import (
	"bytes"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/version"
	"net/http"
	"time"
)

type Client struct {
	Config  *configuration.Configuration
	Group   string
	Manager *Manager
	Context *ClientContext
	Version *version.VersionClient
}

type Storage interface {
	Save(ctx *ClientContext) error
	Load(name string) (*ClientContext, error)
	GetActive() (string, error)
	SetActive(name string) error
	Delete(name string) error
	List() ([]string, error)
}

type FileStorage struct {
	contextDir string // Base directory for contexts
}

type MemoryStorage struct {
	contexts      map[string]*ClientContext // Map of context name to context
	activeContext string                    // Name of active context
}

type Config struct {
	RootDir     string
	APITimeout  time.Duration
	MaxRetries  int
	RetryDelay  time.Duration
	UseInsecure bool
	InMemory    bool
}

type Credentials struct {
	PrivateKey *bytes.Buffer
	Cert       *bytes.Buffer
	Ca         *bytes.Buffer
	CertBundle string
	User       *authentication.User
}

type ClientContext struct {
	Name        string
	Directory   string `json:"-"`
	APIURL      string
	Credentials *Credentials
	client      *http.Client
	config      *Config
}

type Manager struct {
	config *Config
	store  Storage
}

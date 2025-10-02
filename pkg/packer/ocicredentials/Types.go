package ocicredentials

import (
	"bufio"
	"errors"
)

var (
	ErrRegistryRequired   = errors.New("registry URL is required")
	ErrPasswordRequired   = errors.New("password is required when username is provided")
	ErrCredentialsNil     = errors.New("credentials cannot be nil")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnknownConfigKey   = errors.New("unknown credentials key")
	ErrInvalidInputFormat = errors.New("invalid input format")
)

const (
	DefaultRegistry = "registry.simplecontainer.io"
)

const (
	DefaultConfigFile     = ".smrctl/auth.yaml"
	ConfigFilePermissions = 0600
	ConfigDirPermissions  = 0700
)

type Credentials struct {
	Registry string `yaml:"registry" json:"registry"`
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
}

type Manager struct {
	authFilePath string
	Credentials  map[string]*Credentials
}

type InputReader struct {
	reader *bufio.Reader
}

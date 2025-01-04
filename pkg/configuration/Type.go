package configuration

import (
	"github.com/simplecontainer/smr/pkg/node"
)

type Configuration struct {
	Platform       string        `yaml:"platform"`
	OverlayNetwork string        `yaml:"overlaynetwork"`
	Node           string        `yaml:"node"`
	HostPort       HostPort      `yaml:"hostport"`
	KVStore        *KVStore      `yaml:"kvstore"`
	Target         string        `default:"development" yaml:"target"`
	Root           string        `yaml:"root"`
	Certificates   *Certificates `yaml:"certificates"`
	HostHome       string        `yaml:"home"`
	Environment    *Environment  `yaml:"-"`
	Flags          Flags         `yaml:"-"`
}

type HostPort struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type KVStore struct {
	Cluster     []*node.Node `yaml:"cluster"`
	Node        uint64       `yaml:"node"`
	URL         string       `yaml:"url"`
	JoinCluster bool         `yaml:"join"`
}

type Certificates struct {
	Domains *Domains `yaml:"domains"`
	IPs     *IPs     `yaml:"ips"`
}

type IPs struct {
	Members []string `yaml:"members"`
}

type Domains struct {
	Members []string `yaml:"members"`
}

type Flags struct {
	Opt     bool
	Verbose bool
}

type Environment struct {
	HOMEDIR    string
	OPTDIR     string
	PROJECTDIR string
	PROJECT    string
	AGENTIP    string
}

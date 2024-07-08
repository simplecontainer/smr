package container

import (
	"context"
	"github.com/simplecontainer/smr/implementations/container/status"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/utils"
	"strings"
	"sync"
)

type Container struct {
	Static  Static
	Runtime Runtime
	Exports Exports
	Status  status.Status
}

type Static struct {
	Name                   string
	GeneratedName          string
	GeneratedNameNoProject string
	Labels                 map[string]string
	Group                  string
	Image                  string
	Tag                    string
	Replicas               int
	Networks               []string
	Env                    []string
	Entrypoint             []string
	Command                []string
	MappingFiles           []map[string]string
	MappingPorts           []PortMappings
	ExposedPorts           []string
	MountFiles             []string
	Capabilities           []string
	NetworkMode            string
	Privileged             bool
	Readiness              []Readiness
	Definition             v1.Container
}

type Runtime struct {
	Auth               string
	Id                 string
	Networks           map[string]Network
	NetworkLock        sync.RWMutex
	State              string
	FoundRunning       bool
	FirstObserved      bool
	Ready              bool
	Configuration      map[string]string
	Resources          []Resource
	Owner              Owner
	ObjectDependencies []*f.Format
}

type Owner struct {
	Kind            string
	GroupIdentifier string
}

type Network struct {
	NetworkId   string
	NetworkName string
	IP          string
}

type PortMappings struct {
	Container string
	Host      string
}

type Resource struct {
	Identifier string
	Key        string
	Data       map[string]string
	MountPoint string
}

type Exports struct {
	path string
}

type ExecResult struct {
	Stdout string
	Stderr string
	Exit   int
}

// Readiness related

type Readiness struct {
	Name       string
	Operator   string
	Timeout    string
	Body       map[string]string
	Solved     bool
	BodyUnpack map[string]string  `json:"-"`
	Ctx        context.Context    `json:"-"`
	Cancel     context.CancelFunc `json:"-"`
}

type ReadinessState struct {
	Name       string
	Success    bool
	Missing    bool
	Timeout    bool
	Error      error
	TryToSolve bool
	Readiness  *Readiness
}

type ReadinessResult struct {
	Data string
}

// Dependencies related

type ByDepenendecies []*Container

func (d ByDepenendecies) Len() int { return len(d) }
func (d ByDepenendecies) Less(i, j int) bool {
	for _, deps := range d[i].Static.Definition.Spec.Container.Dependencies {
		group, id := utils.ExtractGroupAndId(deps.Name)

		if id == "*" {
			if strings.Contains(d[j].Static.GeneratedNameNoProject, group) {
				return false
			}
		} else {
			if id == d[j].Static.GeneratedNameNoProject {
				return false
			}
		}
	}

	return true
}
func (d ByDepenendecies) Swap(i, j int) { d[i], d[j] = d[j], d[i] }

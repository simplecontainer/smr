package container

import (
	"smr/pkg/definitions"
	"smr/pkg/network"
	"smr/pkg/utils"
	"strings"
)

type Container struct {
	Static  Static
	Runtime Runtime
	Exports Exports
	Status  Status
}

type Static struct {
	Name                   string
	GeneratedName          string
	GeneratedNameNoProject string
	Group                  string
	Image                  string
	Tag                    string
	Replicas               int
	Networks               []string
	Env                    []string
	MappingFiles           []map[string]string
	MappingPorts           []network.PortMappings
	ExposedPorts           []string
	MountFiles             []string
	Definition             definitions.Definition
}

type Runtime struct {
	Auth          string
	Id            string
	Networks      map[string]Network
	State         string
	FoundRunning  bool
	FirstObserved bool
	Ready         bool
	Configuration map[string]any
	Resources     []Resource
}

type Network struct {
	NetworkId string
	IP        string
}

type Status struct {
	DependsSolved  bool
	BackOffRestart bool
	Healthy        bool
	Ready          bool
	Running        bool
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

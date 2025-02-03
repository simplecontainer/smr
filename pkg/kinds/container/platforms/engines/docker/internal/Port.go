package internal

import (
	"fmt"
	"github.com/docker/go-connections/nat"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
)

type Ports struct {
	Ports []*Port
}

type Port struct {
	Container string
	Host      string
}

func NewPorts(ports []v1.ContainerPort) *Ports {
	portsObj := &Ports{
		Ports: make([]*Port, 0),
	}

	for _, port := range ports {
		portsObj.Add(port)
	}

	return portsObj
}

func NewPort(port v1.ContainerPort) *Port {
	return &Port{
		Container: port.Container,
		Host:      port.Host,
	}
}

func (ports *Ports) Add(port v1.ContainerPort) {
	ports.Ports = append(ports.Ports, NewPort(port))
}

func (ports *Ports) ToPortExposed() nat.PortSet {
	NatSet := nat.PortSet{}

	for _, port := range ports.Ports {
		NatSet[nat.Port(port.Container)] = struct{}{}
	}

	return NatSet
}

func (ports *Ports) ToPortMap() nat.PortMap {
	NatMap := nat.PortMap{}

	for _, port := range ports.Ports {
		if port.Host != "" {
			portSpec := fmt.Sprintf("%s:%s", port.Host, port.Container)
			HostPortMapping, err := nat.ParsePortSpec(portSpec)

			if err != nil {
				return NatMap
			}

			var HostPortBinding = make([]nat.PortBinding, 0)

			for _, hpm := range HostPortMapping {
				HostPortBinding = append(HostPortBinding, hpm.Binding)
			}

			NatMap[nat.Port(port.Container)] = HostPortBinding
		}
	}

	return NatMap
}

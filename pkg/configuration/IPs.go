package configuration

import (
	"net"
)

func NewIPs(ips []string) *IPs {
	container := &IPs{Members: []string{}}

	for _, ip := range ips {
		container.Add(ip)
	}

	return container
}

func (ips *IPs) Add(ip string) {
	for _, member := range ips.Members {
		if member == ip {
			return
		}
	}

	if ip == "" {
		return
	}

	ips.Members = append(ips.Members, ip)
}

func (ips *IPs) Remove(ip string) {
	for i, member := range ips.Members {
		if member == ip {
			ips.Members = append(ips.Members[:i], ips.Members[i+1:]...)
		}
	}
}

func (ips *IPs) ToIPNetSlice() []net.IP {
	members := make([]net.IP, 0)

	for _, ip := range ips.Members {
		parsed := net.ParseIP(ip)

		if parsed != nil {
			members = append(members, parsed)
		}
	}

	return members
}

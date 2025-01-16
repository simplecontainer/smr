package network

import "net"

type Nameservers struct {
	NS        []net.IP
	Defaultv4 []net.IP
	Defaultv6 []net.IP
}

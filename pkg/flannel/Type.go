package flannel

import (
	"github.com/flannel-io/flannel/pkg/backend"
	"github.com/flannel-io/flannel/pkg/ip"
	"net"
)

type Flannel struct {
	Backend           string
	FlannelConfigPath string

	Interface          *backend.ExternalInterface
	InterfaceSpecified *net.Interface
	CIDR               []*net.IPNet
	NetMode            int
	Network            ip.IP4Net
	V6Network          ip.IP6Net

	IPv4Enabled bool
	IPv6Enabled bool
	IPv6Masq    bool
}

type Subnet struct {
	PublicIP    string `json:"PublicIP"`
	PublicIPv6  string `json:"PublicIPv6"`
	BackendType string `json:"BackendType"`
	BackendData struct {
		VNI     int    `json:"VNI"`
		VtepMAC string `json:"VtepMAC"`
	} `json:"BackendData"`
}

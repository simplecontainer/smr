package flannel

import (
	"github.com/pkg/errors"
	"net"
	"os"
)

func New(path string) *Flannel {
	return &Flannel{
		Interface:         nil,
		CIDR:              make([]*net.IPNet, 0),
		FlannelConfigPath: path,
	}
}

func (f *Flannel) Clear() error {
	if _, err := os.Stat(f.FlannelConfigPath); err == nil {
		return os.Remove(f.FlannelConfigPath)
	}

	return nil
}

func (f *Flannel) SetBackend(backend string) error {
	switch backend {
	case "wireguard", "vxlan":
		f.Backend = backend
		return nil
	default:
		return errors.Errorf("invalid backend: %s", backend)
	}
}
func (f *Flannel) SetCIDR(cidr string) error {
	_, tmp, err := net.ParseCIDR(cidr)

	if err != nil {
		return err
	}

	f.CIDR = append(f.CIDR, tmp)

	f.NetMode, err = findNetMode(f.CIDR)

	if err != nil {
		return errors.Wrap(err, "failed to check netMode for flannel")
	}

	return nil
}
func (f *Flannel) SetInterface(name string) error {
	if name == "" {
		return nil
	}

	iface, err := net.InterfaceByName(name)

	if err != nil {
		return err
	}

	f.InterfaceSpecified = iface
	return nil
}
func (f *Flannel) EnableIPv6(enable bool) error {
	if enable {
		if _, err := os.Stat("/proc/sys/net/bridge/bridge-nf-call-iptables"); os.IsNotExist(err) {
			return err
		}

		f.IPv4Enabled = true
	}

	return nil
}
func (f *Flannel) EnableIPv4(enable bool) error {
	if enable {
		if _, err := os.Stat("/proc/sys/net/bridge/bridge-nf-call-ip6tables"); os.IsNotExist(err) {
			return err
		}

		f.IPv4Enabled = true
	}

	return nil
}
func (f *Flannel) MaskIPv6(enabled bool) {
	f.IPv6Masq = enabled
}

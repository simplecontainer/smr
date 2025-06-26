package nameservers

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func New() *Nameservers {
	return &Nameservers{
		NS: make([]net.IP, 0),
	}
}

func (nameservers *Nameservers) Add(ip net.IP) {
	if !ip.IsLoopback() {
		nameservers.NS = append(nameservers.NS, ip)
	}
}

func (nameservers *Nameservers) ToString() []string {
	tmp := make([]string, 0)

	for _, ns := range nameservers.NS {
		if ns.String() != "" {
			tmp = append(tmp, ns.String())
		}
	}

	return tmp
}

func NewfromResolvConf(ipv6 bool) (*Nameservers, []string, error) {
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return nil, nil, fmt.Errorf("could not open /etc/resolv.conf: %w", err)
	}
	defer file.Close()

	var nameservers = New()
	var search = []string{}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "nameserver") {
			parts := strings.Fields(line)

			if len(parts) > 1 {
				nameservers.Add(net.ParseIP(parts[1]))
			}
		}

		if strings.HasPrefix(line, "search") {
			parts := strings.Fields(line)

			if len(parts) > 1 {
				search = append(search, parts[1:]...)
			}
		}
	}

	if len(nameservers.NS) == 0 {
		if ipv6 {
			nameservers.NS = GetDefaultNSv6()
		} else {
			nameservers.NS = GetDefaultNSv4()
		}
	}

	return nameservers, search, nil
}

func GetDefaultNSv4() []net.IP {
	return []net.IP{
		net.ParseIP("8.8.8.8"),
		net.ParseIP("4.4.4.4"),
	}
}

func GetDefaultNSv6() []net.IP {
	return []net.IP{
		net.ParseIP("2001:4860:4860::8888"),
		net.ParseIP("2001:4860:4860::8844"),
	}
}

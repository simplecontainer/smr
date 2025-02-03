package ips

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func New() *IPs {
	return &IPs{
		IPs: make([]net.IP, 0),
	}
}

func (ips *IPs) Add(ip net.IP) {
	ips.IPs = append(ips.IPs, ip)
}

func NewfromEtcHosts() (*IPs, error) {
	file, err := os.Open("/etc/hosts")
	if err != nil {
		return nil, fmt.Errorf("could not open /etc/hosts: %w", err)
	}
	defer file.Close()

	var IPs = New()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.Fields(line)

		if len(parts) > 1 {
			IPs.Add(net.ParseIP(parts[0]))
		}

	}

	return IPs, nil
}

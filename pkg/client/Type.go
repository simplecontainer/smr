package client

import (
	"net"
	"net/http"
)

type Http struct {
	Clients map[string]*Client
}

type Client struct {
	Http     *http.Client
	Username string
	API      string
	IPs      []net.IP
	Domains  []string
}

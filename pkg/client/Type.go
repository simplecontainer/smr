package client

import "net/http"

type Http struct {
	Clients map[string]*Client
}

type Client struct {
	Http *http.Client
	API  string
}

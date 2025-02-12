package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/cluster"
	"github.com/simplecontainer/smr/pkg/keys"
	"net"
	"net/http"
	"net/url"
)

func NewHttpClients() *Http {
	return &Http{Clients: make(map[string]*Client)}
}

func (http *Http) Get(username string) *Client {
	return http.Clients[username]
}

func (http *Http) Append(username string, client *Client) {
	http.Clients[username] = client
}

func (http *Http) FindValidFor(DomainOrIp string) *Client {
	u, err := url.Parse(DomainOrIp)

	if err != nil {
		return nil
	}

	if ip := net.ParseIP(u.Hostname()); ip != nil {
		for _, h := range http.Clients {
			for _, i := range h.IPs {
				if i.Equal(ip) {
					return h
				}
			}
		}
	} else {
		for _, h := range http.Clients {
			for _, d := range h.Domains {
				if d == u.Hostname() {
					return h
				}
			}
		}
	}

	fmt.Println("found 0 results")
	return nil
}

func GenerateHttpClients(nodeName string, keys *keys.Keys, cluster *cluster.Cluster) (*Http, error) {
	hc := NewHttpClients()

	// Configure custom users
	for username, c := range keys.Clients {
		httpClient, err := GenerateHttpClient(keys.CA, c)

		if err != nil {
			return nil, err
		}

		hc.Append(username, &Client{
			Http:     httpClient,
			Username: username,
			API:      fmt.Sprintf("%s:%s", c.Certificate.DNSNames[0], "1443"),
			Domains:  c.Certificate.DNSNames,
			IPs:      c.Certificate.IPAddresses,
		})
	}

	// Configure node users
	if cluster != nil && cluster.Cluster != nil {
		for _, c := range cluster.Cluster.Nodes {
			_, ok := keys.Clients[cluster.Node.NodeName]

			if ok {
				httpClient, err := GenerateHttpClient(keys.CA, keys.Clients[cluster.Node.NodeName])

				if err != nil {
					return nil, err
				}

				hc.Append(c.NodeName, &Client{
					Http:     httpClient,
					Username: c.NodeName,
					API:      c.API,
					Domains:  keys.Clients[cluster.Node.NodeName].Certificate.DNSNames,
					IPs:      keys.Clients[cluster.Node.NodeName].Certificate.IPAddresses,
				})
			} else {
				return nil, errors.New("certificates for the node missing")
			}
		}
	}

	return hc, nil
}

func GenerateHttpClient(ca *keys.CA, client *keys.Client) (*http.Client, error) {
	var PEMCertificate []byte = make([]byte, 0)
	var PEMPrivateKey []byte = make([]byte, 0)

	var err error

	PEMCertificate, err = keys.PEMEncode(keys.CERTIFICATE, client.CertificateBytes)
	PEMPrivateKey, err = keys.PEMEncode(keys.PRIVATE_KEY, client.PrivateKeyBytes)

	cert, err := tls.X509KeyPair(PEMCertificate, PEMPrivateKey)
	if err != nil {
		return nil, err
	}

	CAPool := x509.NewCertPool()
	CAPool.AddCert(ca.Certificate)

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      CAPool,
				Certificates: []tls.Certificate{cert},
			},
		},
	}, nil
}

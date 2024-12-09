package client

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/simplecontainer/smr/pkg/keys"
	"net/http"
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

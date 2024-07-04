package manager

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/simplecontainer/smr/pkg/keys"
	"net/http"
)

func GenerateHttpClient(keys *keys.Keys) (*http.Client, error) {
	cert, err := tls.X509KeyPair(keys.ClientCertPem.Bytes(), keys.ClientPrivateKey.Bytes())
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(keys.CAPem.Bytes())

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert},
			},
		},
	}, nil
}

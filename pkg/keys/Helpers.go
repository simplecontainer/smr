package keys

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"
)

func (keys *Keys) GenerateHttpClient() (*http.Client, error) {
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

func (keys *Keys) GeneratePemBundle() string {
	return fmt.Sprintf("%s\n%s\n%s\n", strings.TrimSpace(keys.ClientPrivateKey.String()), strings.TrimSpace(keys.ClientCertPem.String()), strings.TrimSpace(keys.CAPem.String()))
}

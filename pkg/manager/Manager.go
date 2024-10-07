package manager

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/simplecontainer/smr/pkg/keys"
	"net/http"
)

func GenerateHttpClient(k *keys.Keys) (*http.Client, error) {
	var PEMCertificate []byte
	var PEMPrivateKey []byte

	var err error

	PEMCertificate, err = keys.PEMEncode(keys.CERTIFICATE, k.Server.CertificateBytes)
	PEMPrivateKey, err = keys.PEMEncode(keys.PRIVATE_KEY, k.Server.PrivateKeyBytes)

	cert, err := tls.X509KeyPair(PEMCertificate, PEMPrivateKey)
	if err != nil {
		return nil, err
	}

	CAPool := x509.NewCertPool()
	CAPool.AddCert(k.CA.Certificate)

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      CAPool,
				Certificates: []tls.Certificate{cert},
			},
		},
	}, nil
}

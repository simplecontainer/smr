package keys

import (
	"crypto/ecdsa"
	"crypto/x509"
)

type Keys struct {
	CA       *CA
	Server   *Server
	Clients  map[string]*Client
	Reloader *keypairReloader
}

type CA struct {
	PrivateKey       *ecdsa.PrivateKey
	Certificate      *x509.Certificate
	PrivateKeyPath   string
	CertificatePath  string
	CertificateBytes []byte
	PrivateKeyBytes  []byte
	Sni              int64
}

type Server struct {
	PrivateKey       *ecdsa.PrivateKey
	Certificate      *x509.Certificate
	PrivateKeyPath   string
	CertificatePath  string
	CertificateBytes []byte
	PrivateKeyBytes  []byte
	Sni              int64
}

type Client struct {
	PrivateKey       *ecdsa.PrivateKey
	Certificate      *x509.Certificate
	PrivateKeyPath   string
	CertificatePath  string
	CertificateBytes []byte
	PrivateKeyBytes  []byte
	Sni              int64
}

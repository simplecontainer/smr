package keys

import (
	"crypto/rsa"
	"crypto/x509"
)

type Keys struct {
	CA       *CA
	Server   *Server
	Clients  map[string]*Client
	Reloader *keypairReloader
}

type CA struct {
	PrivateKey       *rsa.PrivateKey
	Certificate      *x509.Certificate
	PrivateKeyPath   string
	CertificatePath  string
	CertificateBytes []byte
	PrivateKeyBytes  []byte
	Sni              int64
}

type Server struct {
	PrivateKey       *rsa.PrivateKey
	Certificate      *x509.Certificate
	PrivateKeyPath   string
	CertificatePath  string
	CertificateBytes []byte
	PrivateKeyBytes  []byte
	Sni              int64
}

type Client struct {
	PrivateKey       *rsa.PrivateKey
	Certificate      *x509.Certificate
	PrivateKeyPath   string
	CertificatePath  string
	CertificateBytes []byte
	PrivateKeyBytes  []byte
	Sni              int64
}

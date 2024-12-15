package keys

import (
	"crypto/ecdsa"
	"crypto/x509"
)

type Keys struct {
	CA       *CA
	Server   *Server
	Clients  map[string]*Client
	Reloader *keypairReloader `json:"-"`
}

type Encrypted struct {
	Keys string
}

type CA struct {
	PrivateKey       *ecdsa.PrivateKey `json:"-"`
	Certificate      *x509.Certificate `json:"-"`
	PrivateKeyPath   string            `json:"-"`
	CertificatePath  string            `json:"-"`
	CertificateBytes []byte
	PrivateKeyBytes  []byte
	Sni              int64
}

type Server struct {
	PrivateKey       *ecdsa.PrivateKey `json:"-"`
	Certificate      *x509.Certificate `json:"-"`
	PrivateKeyPath   string            `json:"-"`
	CertificatePath  string            `json:"-"`
	CertificateBytes []byte
	PrivateKeyBytes  []byte
	Sni              int64
}

type Client struct {
	PrivateKey       *ecdsa.PrivateKey `json:"-"`
	Certificate      *x509.Certificate `json:"-"`
	PrivateKeyPath   string            `json:"-"`
	CertificatePath  string            `json:"-"`
	CertificateBytes []byte
	PrivateKeyBytes  []byte
	Sni              int64
}

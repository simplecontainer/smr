package keys

import (
	"crypto/ecdsa"
	"crypto/x509"
	"math/big"
)

type Keys struct {
	CA       *CA
	Server   *Server
	Clients  map[string]*Client
	Reloader *keypairReloader `json:"-"`
	Sni      uint64
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
	Sni              *big.Int
}

type Server struct {
	PrivateKey       *ecdsa.PrivateKey `json:"-"`
	Certificate      *x509.Certificate `json:"-"`
	PrivateKeyPath   string            `json:"-"`
	CertificatePath  string            `json:"-"`
	CertificateBytes []byte
	PrivateKeyBytes  []byte
	Sni              *big.Int
}

type Client struct {
	PrivateKey       *ecdsa.PrivateKey `json:"-"`
	Certificate      *x509.Certificate `json:"-"`
	PrivateKeyPath   string            `json:"-"`
	CertificatePath  string            `json:"-"`
	CertificateBytes []byte
	PrivateKeyBytes  []byte
	Sni              *big.Int
}

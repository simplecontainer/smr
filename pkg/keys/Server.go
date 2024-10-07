package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"github.com/simplecontainer/smr/pkg/static"
	"math/big"
	"net"
	"os"
	"time"
)

func NewServer() *Server {
	return &Server{
		Sni: 0,
	}
}

func (server *Server) Generate(ca *CA, domains []string, ips []net.IP, CN string) error {
	var err error

	server.Sni = server.Sni + 1

	server.PrivateKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	server.PrivateKeyBytes, err = x509.MarshalPKCS8PrivateKey(server.PrivateKey)

	if err != nil {
		return err
	}

	var PublicKey []byte
	PublicKey, err = x509.MarshalPKIXPublicKey(&server.PrivateKey.PublicKey)

	if err != nil {
		return err
	}

	SubjectKeyIdentifier := sha1.Sum(PublicKey)

	server.Certificate = &x509.Certificate{
		SerialNumber: big.NewInt(server.Sni),
		Subject: pkix.Name{
			Organization: []string{"simplecontainer"},
			CommonName:   CN,
		},
		DNSNames:     append(domains, []string{fmt.Sprintf("smr-agent.%s", static.SMR_LOCAL_DOMAIN)}...),
		IPAddresses:  append(ips, net.IPv6loopback),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: SubjectKeyIdentifier[:],
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}

	server.CertificateBytes, err = x509.CreateCertificate(rand.Reader, server.Certificate, ca.Certificate, &server.PrivateKey.PublicKey, ca.PrivateKey)

	if err != nil {
		return err
	}

	return nil
}
func (server *Server) Write(directory string) error {
	err := os.MkdirAll(directory, os.ModePerm)

	if err != nil {
		panic(err)
	}

	PemCertificate, err := PEMEncode(CERTIFICATE, server.CertificateBytes)

	if err != nil {
		return err
	}

	PemPrivateKey, err := PEMEncode(PRIVATE_KEY, server.PrivateKeyBytes)

	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s/server.crt", directory), PemCertificate, 0644)

	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s/server.key", directory), PemPrivateKey, 0644)

	if err != nil {
		return err
	}

	return nil
}
func (server *Server) Read(directory string) error {
	PemCertificate, err := os.ReadFile(fmt.Sprintf("%s/server.crt", directory))
	if err != nil {
		return err
	}

	PemPrivateKey, err := os.ReadFile(fmt.Sprintf("%s/server.key", directory))
	if err != nil {
		return err
	}

	server.CertificateBytes = PEMDecode(PemCertificate)
	server.Certificate, err = x509.ParseCertificate(PEMDecode(PemCertificate))

	if err != nil {
		return err
	}

	var PrivateKeyTmp any
	PrivateKeyTmp, err = x509.ParsePKCS8PrivateKey(PEMDecode(PemPrivateKey))

	if err != nil {
		return err
	}

	server.PrivateKey = PrivateKeyTmp.(*rsa.PrivateKey)

	server.PrivateKeyBytes, err = x509.MarshalPKCS8PrivateKey(server.PrivateKey)

	if err != nil {
		return err
	}

	return nil
}

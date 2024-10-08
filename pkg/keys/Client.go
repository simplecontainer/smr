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

func NewClient() *Client {
	return &Client{
		Sni: 0,
	}
}

func (client *Client) Generate(ca *CA, domains []string, ips []net.IP, CN string) error {
	var err error

	client.Sni = client.Sni + 1

	client.PrivateKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	client.PrivateKeyBytes, err = x509.MarshalPKCS8PrivateKey(client.PrivateKey)

	if err != nil {
		return err
	}

	var PublicKey []byte
	PublicKey, err = x509.MarshalPKIXPublicKey(&client.PrivateKey.PublicKey)
	if err != nil {
		return err
	}

	SubjectKeyIdentifier := sha1.Sum(PublicKey)

	client.Certificate = &x509.Certificate{
		SerialNumber: big.NewInt(client.Sni),
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

	client.CertificateBytes, err = x509.CreateCertificate(rand.Reader, client.Certificate, ca.Certificate, &client.PrivateKey.PublicKey, ca.PrivateKey)

	if err != nil {
		return err
	}

	client.Certificate, err = x509.ParseCertificate(client.CertificateBytes)

	if err != nil {
		return err
	}

	return nil
}
func (client *Client) Write(directory string, username string) error {
	err := os.MkdirAll(directory, os.ModePerm)

	if err != nil {
		panic(err)
	}

	PemCertificate, err := PEMEncode(CERTIFICATE, client.CertificateBytes)

	if err != nil {
		return err
	}

	PemPrivateKey, err := PEMEncode(PRIVATE_KEY, client.PrivateKeyBytes)

	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s/%s.crt", directory, username), PemCertificate, 0644)

	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s/%s.key", directory, username), PemPrivateKey, 0644)

	if err != nil {
		return err
	}

	return nil
}
func (client *Client) Read(directory string, username string) error {
	PemCertificate, err := os.ReadFile(fmt.Sprintf("%s/%s.crt", directory, username))
	if err != nil {
		return err
	}

	PemPrivateKey, err := os.ReadFile(fmt.Sprintf("%s/%s.key", directory, username))
	if err != nil {
		return err
	}

	client.CertificateBytes = PEMDecode(PemCertificate)
	client.Certificate, err = x509.ParseCertificate(PEMDecode(PemCertificate))

	if err != nil {
		return err
	}

	var PrivateKeyTmp any
	PrivateKeyTmp, err = x509.ParsePKCS8PrivateKey(PEMDecode(PemPrivateKey))

	if err != nil {
		return err
	}

	client.PrivateKey = PrivateKeyTmp.(*rsa.PrivateKey)

	client.PrivateKeyBytes, err = x509.MarshalPKCS8PrivateKey(client.PrivateKey)

	if err != nil {
		return err
	}

	return nil
}

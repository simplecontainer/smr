package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"os"
	"time"
)

func NewCA() *CA {
	return &CA{
		Sni: 0,
	}
}

func (ca *CA) Generate() error {
	ca.Sni = ca.Sni + 1

	ca.Certificate = &x509.Certificate{
		SerialNumber: big.NewInt(ca.Sni),
		Subject: pkix.Name{
			Organization:  []string{"simplecontainer"},
			Country:       []string{"BA"},
			Province:      []string{""},
			Locality:      []string{"Zivinice"},
			StreetAddress: []string{"BB"},
			PostalCode:    []string{"75270"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	var err error

	ca.PrivateKey, err = rsa.GenerateKey(rand.Reader, 4096)

	if err != nil {
		return err
	}

	ca.PrivateKeyBytes, err = x509.MarshalPKCS8PrivateKey(ca.PrivateKey)

	if err != nil {
		return err
	}

	ca.CertificateBytes, err = x509.CreateCertificate(rand.Reader, ca.Certificate, ca.Certificate, &ca.PrivateKey.PublicKey, ca.PrivateKey)

	if err != nil {
		return err
	}

	ca.Certificate, err = x509.ParseCertificate(ca.CertificateBytes)

	if err != nil {
		return err
	}

	return nil
}
func (ca *CA) Write(directory string) error {
	err := os.MkdirAll(directory, os.ModePerm)

	if err != nil {
		panic(err)
	}

	PemCertificate, err := PEMEncode(CERTIFICATE, ca.CertificateBytes)

	if err != nil {
		return err
	}

	PemPrivateKey, err := PEMEncode(PRIVATE_KEY, ca.PrivateKeyBytes)

	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s/ca.crt", directory), PemCertificate, 0644)

	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("%s/ca.key", directory), PemPrivateKey, 0644)

	if err != nil {
		return err
	}

	return nil
}
func (ca *CA) Read(directory string) error {
	PemCertificate, err := os.ReadFile(fmt.Sprintf("%s/ca.crt", directory))
	if err != nil {
		return err
	}

	PemPrivateKey, err := os.ReadFile(fmt.Sprintf("%s/ca.key", directory))
	if err != nil {
		return err
	}

	ca.CertificateBytes = PEMDecode(PemCertificate)
	ca.Certificate, err = x509.ParseCertificate(PEMDecode(PemCertificate))

	if err != nil {
		return err
	}

	var PrivateKeyTmp any
	PrivateKeyTmp, err = x509.ParsePKCS8PrivateKey(PEMDecode(PemPrivateKey))

	if err != nil {
		return err
	}

	ca.PrivateKey = PrivateKeyTmp.(*rsa.PrivateKey)

	ca.PrivateKeyBytes, err = x509.MarshalPKCS8PrivateKey(ca.PrivateKey)

	if err != nil {
		return err
	}

	return nil
}

package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"os"
	"time"
)

func NewCA() *CA {
	return &CA{
		Sni: nil,
	}
}

func (ca *CA) Generate() error {
	ca.Sni = generateSerialNumber()

	ca.Certificate = &x509.Certificate{
		SerialNumber: ca.Sni,
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

	ca.PrivateKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

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

	ca.CertificatePath = fmt.Sprintf("%s/ca.crt", directory)
	ca.PrivateKeyPath = fmt.Sprintf("%s/ca.key", directory)

	err = os.WriteFile(ca.CertificatePath, PemCertificate, 0644)

	if err != nil {
		return err
	}

	err = os.WriteFile(ca.PrivateKeyPath, PemPrivateKey, 0644)

	if err != nil {
		return err
	}

	return nil
}
func (ca *CA) Read(directory string) error {
	ca.CertificatePath = fmt.Sprintf("%s/ca.crt", directory)
	ca.PrivateKeyPath = fmt.Sprintf("%s/ca.key", directory)

	PemCertificate, err := os.ReadFile(ca.CertificatePath)
	if err != nil {
		return err
	}

	PemPrivateKey, err := os.ReadFile(ca.PrivateKeyPath)
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

	ca.PrivateKey = PrivateKeyTmp.(*ecdsa.PrivateKey)

	ca.PrivateKeyBytes, err = x509.MarshalPKCS8PrivateKey(ca.PrivateKey)

	if err != nil {
		return err
	}

	return nil
}

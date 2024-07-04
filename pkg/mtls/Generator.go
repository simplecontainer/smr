package mtls

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/keys"
	"math/big"
	"net"
	"time"
)

/*
Taken and modified from the: https://gist.github.com/shaneutt/5e1995295cff6721c89a71d13a71c251
*/

func GenerateKeys(keys *keys.Keys, config *configuration.Configuration) error {
	// set up our CA certificate
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{"Simple container manager."},
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
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// create our private and public key
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	// pem encode
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	keys.CAPem = caPEM

	caPrivKeyPEMPKCS8, err := x509.MarshalPKCS8PrivateKey(caPrivKey)

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: caPrivKeyPEMPKCS8,
	})

	keys.CAPrivateKey = caPrivKeyPEM

	keys.ServerPrivateKey, keys.ServerCertPem, err = generateCertPrivKeyPair(config, ca, caPrivKey)

	if err != nil {
		return err
	}

	keys.ClientPrivateKey, keys.ClientCertPem, err = generateCertPrivKeyPair(config, ca, caPrivKey)

	if err != nil {
		return err
	}

	return nil
}

func generateCertPrivKeyPair(config *configuration.Configuration, ca *x509.Certificate, caPrivKey *rsa.PrivateKey) (*bytes.Buffer, *bytes.Buffer, error) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{"Simple container manager."},
			Country:       []string{"BA"},
			Province:      []string{""},
			Locality:      []string{"Zivinice"},
			StreetAddress: []string{"BB"},
			PostalCode:    []string{"75270"},
		},
		DNSNames:     []string{config.Flags.DaemonDomain, "smr-agent"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPKCS8, err := x509.MarshalPKCS8PrivateKey(certPrivKey)

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: certPrivKeyPKCS8,
	})

	return certPrivKeyPEM, certPEM, nil
}

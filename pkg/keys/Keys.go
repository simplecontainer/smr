package keys

import (
	"crypto/rand"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/configuration"
	"io/fs"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
)

func NewKeys() *Keys {
	return &Keys{
		CA:      NewCA(),
		Server:  NewServer(),
		Clients: NewClients(),
		Sni:     0,
	}
}

func NewClients() map[string]*Client {
	return make(map[string]*Client)
}

func (keys *Keys) AppendClient(username string, newClient *Client) {
	keys.Clients[username] = newClient
}

func (keys *Keys) GenerateCA() error {
	return keys.CA.Generate()
}

func (keys *Keys) GenerateServer(domains *configuration.Domains, ips *configuration.IPs) error {
	hostname, err := os.Hostname()

	if err != nil {
		hostname = "simplecontainer"
	}

	return keys.Server.Generate(keys.CA, domains, ips, hostname)
}

func (keys *Keys) GenerateClient(domains *configuration.Domains, ips *configuration.IPs, username string) error {
	keys.Clients[username] = NewClient()
	return keys.Clients[username].Generate(keys.CA, domains, ips, username)
}

func (keys *Keys) CAExists(directory string, username string) error {
	return keys.CA.Read(directory)
}

func (keys *Keys) ServerExists(directory string, agent string) error {
	return keys.Server.Read(directory, agent)
}

func (keys *Keys) ClientExists(directory string, username string) error {
	var usernameCert = fmt.Sprintf("%s.pem", username)

	return filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		for _, s := range []string{"pem"} {
			if s == usernameCert {
				return errors.New("username already exists")
			}
		}

		return nil
	})
}

func (keys *Keys) LoadClients(directory string) error {
	var files []string
	err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		for _, s := range []string{"pem"} {
			if strings.HasSuffix(path, "."+s) {
				basename := filepath.Base(path)
				basename = strings.TrimSuffix(basename, filepath.Ext(basename))

				files = append(files, basename)
				return nil
			}
		}

		return nil
	})

	for _, username := range files {
		keys.Clients[username] = NewClient()
		err = keys.Clients[username].Read(directory, username)

		if err != nil {
			return err
		}
	}

	return nil
}
func (keys *Keys) LoadClient(username string, bundle string) error {
	keys.Clients[username] = NewClient()

	certificate, privateKey, err := ParsePemBundle(bundle)

	if err != nil {
		return err
	}

	return keys.Clients[username].Load(certificate, privateKey)
}

func (keys *Keys) GeneratePemBundle(directory string, username string, client *Client) error {
	var PemCertificateClient []byte
	var PemPrivateKeyClient []byte
	var PemCertificateCA []byte

	var err error

	PemCertificateClient, err = PEMEncode(CERTIFICATE, client.CertificateBytes)

	if err != nil {
		return err
	}

	PemPrivateKeyClient, err = PEMEncode(PRIVATE_KEY, client.PrivateKeyBytes)

	if err != nil {
		return err
	}

	PemCertificateCA, err = PEMEncode(CERTIFICATE, keys.CA.CertificateBytes)

	if err != nil {
		return err
	}

	bundle := fmt.Sprintf("%s\n%s\n%s\n", strings.TrimSpace(string(PemPrivateKeyClient)), strings.TrimSpace(string(PemCertificateClient)), strings.TrimSpace(string(PemCertificateCA)))

	err = os.WriteFile(fmt.Sprintf("%s/%s.pem", directory, username), []byte(bundle), 0644)
	if err != nil {
		return err
	}

	return nil
}

func ParsePemBundle(bundle string) ([]byte, []byte, error) {
	blocks, err := PEMParse(bundle)

	if err != nil {
		return nil, nil, err
	}

	var certificate []byte
	var privateKey []byte

	for _, block := range blocks {
		parsed, _ := pem.Decode([]byte(block))

		switch parsed.Type {
		case CERTIFICATE:
			var isCA bool
			isCA, err = IsCA(parsed)

			if err != nil {
				return nil, nil, err
			}

			if !isCA {
				certificate = parsed.Bytes
			}
			break
		case PRIVATE_KEY:
			privateKey = parsed.Bytes
			break
		}
	}

	if len(certificate) == 0 || len(privateKey) == 0 {
		return nil, nil, errors.New("invalid certificate or private key")
	}

	return certificate, privateKey, err
}

func generateSerialNumber() *big.Int {
	// Generate a random serial number as a large integer (64-bit in this case)
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128)) // 128-bit serial number
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}
	return serialNumber
}

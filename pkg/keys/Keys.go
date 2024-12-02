package keys

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func NewKeys() *Keys {
	return &Keys{
		CA:      NewCA(),
		Server:  NewServer(),
		Clients: NewClients(),
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

func (keys *Keys) GenerateServer(domains []string, ips []string) error {
	hostname, err := os.Hostname()

	if err != nil {
		hostname = "simplecontainer"
	}

	var ip []net.IP = make([]net.IP, 0)

	for _, IP := range ips {
		ip = append(ip, net.ParseIP(IP))
	}

	return keys.Server.Generate(keys.CA, domains, ip, hostname)
}

func (keys *Keys) GenerateClient(domains []string, ips []string, username string) error {
	var ip []net.IP = make([]net.IP, 0)

	for _, IP := range ips {
		ip = append(ip, net.ParseIP(IP))
	}

	keys.Clients[username] = NewClient()
	return keys.Clients[username].Generate(keys.CA, domains, ip, username)
}

func (keys *Keys) CAExists(directory string, username string) error {
	return keys.CA.Read(directory)
}

func (keys *Keys) ServerExists(directory string, username string) error {
	return keys.Server.Read(directory)
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

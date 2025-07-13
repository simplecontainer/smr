package internal

import (
	"encoding/base64"
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"net"
	"os"
	"strings"
)

type Auth struct {
	Auth transport.AuthMethod
}

func NewAuth() *Auth {
	return &Auth{
		Auth: nil,
	}
}

func (auth *Auth) Http(definition *v1.HttpAuthDefinition) error {
	auth.Auth = &http.BasicAuth{
		Username: definition.Spec.Username,
		Password: definition.Spec.Password,
	}

	return nil
}

func (auth *Auth) Ssh(definition *v1.CertKeyDefinition) error {
	b64decoded, _ := base64.StdEncoding.DecodeString(definition.Spec.PrivateKey)
	tmp, err := ssh.NewPublicKeys(ssh.DefaultUsername, b64decoded, definition.Spec.PrivateKeyPassword)

	if err != nil {
		return err
	}

	knownHostsFile, err := createKnownHostsFile()
	if err != nil {
		return err
	}

	tmp.HostKeyCallback = func(hostname string, remote net.Addr, key gossh.PublicKey) error {
		// Try the known_hosts file first using the knownhosts package
		callback, err := knownhosts.New(knownHostsFile)
		if err != nil {
			return err
		}

		err = callback(hostname, remote, key)
		if err != nil {
			if strings.Contains(err.Error(), "key is unknown") {
				logger.Log.Info("Adding new host key", zap.String("hostname", hostname))
				err = addHostKey(knownHostsFile, hostname, key)
				if err != nil {
					return err
				}

				return nil
			}
			return err
		}

		return nil
	}

	auth.Auth = tmp
	return nil
}

func createKnownHostsFile() (string, error) {
	sshDir := "/home/node/.ssh"
	knownHosts := "/home/node/.ssh/known_hosts"

	// Create .ssh directory if it doesn't exist
	if _, err := os.Stat(sshDir); os.IsNotExist(err) {
		err := os.MkdirAll(sshDir, 0700)
		if err != nil {
			return "", err
		}
	}

	// Create known_hosts file if it doesn't exist
	if _, err := os.Stat(knownHosts); os.IsNotExist(err) {
		file, err := os.Create(knownHosts)
		if err != nil {
			return "", err
		}
		defer file.Close()

		// Set proper permissions
		err = os.Chmod(knownHosts, 0600)
		if err != nil {
			return "", err
		}
	}

	return knownHosts, nil
}

func addHostKey(knownHostsFile, hostname string, key gossh.PublicKey) error {
	file, err := os.OpenFile(knownHostsFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	keyString := gossh.MarshalAuthorizedKey(key)
	line := fmt.Sprintf("%s %s", hostname, strings.TrimSpace(string(keyString)))

	_, err = file.WriteString(line + "\n")
	return err
}

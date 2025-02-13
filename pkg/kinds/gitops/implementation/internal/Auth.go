package internal

import (
	"encoding/base64"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"os"
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

	err = createKnownHostsFile()

	if err != nil {
		return err
	}

	tmp.HostKeyCallback, err = ssh.NewKnownHostsCallback("/home/node/.ssh/known_hosts")

	if err != nil {
		return err
	}

	auth.Auth = tmp
	return nil
}

func createKnownHostsFile() error {
	known_hosts := "/home/node/.ssh/known_hosts" // specify the file path

	if _, err := os.Stat(known_hosts); os.IsNotExist(err) {
		file, err := os.Create(known_hosts)

		if err != nil {
			return err
		}

		defer file.Close()
		return nil
	} else {
		return nil
	}
}

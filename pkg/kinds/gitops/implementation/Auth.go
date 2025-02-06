package implementation

import (
	"encoding/base64"
	"errors"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"os"
)

func (gitops *Gitops) GenerateSshAuth(definition *v1.CertKeyDefinition) (transport.AuthMethod, error) {
	b64decoded, _ := base64.StdEncoding.DecodeString(definition.Spec.PrivateKey)

	auth, err := ssh.NewPublicKeys(ssh.DefaultUsername, b64decoded, definition.Spec.PrivateKeyPassword)

	if err != nil {
		return nil, err
	}

	if generateSshKnownHosts() != nil {
		return nil, errors.New("failed to generate ssh known_hosts file")
	}

	auth.HostKeyCallback, err = ssh.NewKnownHostsCallback("/home/node/.ssh/known_hosts")

	return auth, nil
}

func (gitops *Gitops) GenerateHttpAuth(definition *v1.HttpAuthDefinition) (transport.AuthMethod, error) {
	return &http.BasicAuth{
		Username: definition.Spec.Username,
		Password: definition.Spec.Password,
	}, nil
}

func generateSshKnownHosts() error {
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

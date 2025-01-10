package implementation

import (
	"encoding/base64"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
)

func (gitops *Gitops) GenerateSshAuth(definition *v1.CertKeyDefinition) (transport.AuthMethod, error) {
	b64decoded, _ := base64.StdEncoding.DecodeString(definition.Spec.PrivateKey)

	auth, err := ssh.NewPublicKeys(ssh.DefaultUsername, b64decoded, definition.Spec.PrivateKeyPassword)

	if err != nil {
		return nil, err
	}

	auth.HostKeyCallback, err = ssh.NewKnownHostsCallback()

	return auth, nil
}

func (gitops *Gitops) GenerateHttpAuth(definition *v1.HttpAuthDefinition) (transport.AuthMethod, error) {
	return &http.BasicAuth{
		Username: definition.Spec.Username,
		Password: definition.Spec.Password,
	}, nil
}

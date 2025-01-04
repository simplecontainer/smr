package implementation

import (
	"encoding/base64"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func (gitops *Gitops) GetAuth() (transport.AuthMethod, error) {
	if gitops.AuthInternal.HttpAuth != nil {
		return &gitHttp.BasicAuth{
			Username: gitops.AuthInternal.HttpAuth.Username,
			Password: gitops.AuthInternal.HttpAuth.Password,
		}, nil
	}

	if gitops.AuthInternal.CertKey != nil {
		b64decoded, _ := base64.StdEncoding.DecodeString(gitops.AuthInternal.CertKey.PrivateKey)
		auth, err := ssh.NewPublicKeys(ssh.DefaultUsername, b64decoded, gitops.AuthInternal.CertKey.PrivateKeyPassword)
		auth.HostKeyCallback, err = ssh.NewKnownHostsCallback()

		if err != nil {
			return nil, err
		}

		return auth, nil
	}

	return nil, nil
}

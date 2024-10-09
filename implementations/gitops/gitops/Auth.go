package gitops

import (
	"encoding/base64"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func (gitops *Gitops) GetAuth() (transport.AuthMethod, error) {
	var auth transport.AuthMethod = nil
	var err error = nil

	if gitops.AuthInternal.HttpAuth != nil {
		auth = &gitHttp.BasicAuth{
			Username: gitops.AuthInternal.HttpAuth.Username,
			Password: gitops.AuthInternal.HttpAuth.Password,
		}
	}

	if gitops.AuthInternal.CertKey != nil {
		b64decoded, _ := base64.StdEncoding.DecodeString(gitops.AuthInternal.CertKey.PrivateKey)
		auth, err = ssh.NewPublicKeys(ssh.DefaultUsername, b64decoded, gitops.AuthInternal.CertKey.PrivateKeyPassword)
	}

	return auth, err
}

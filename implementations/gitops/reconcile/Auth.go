package reconcile

import (
	"encoding/base64"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/simplecontainer/smr/implementations/gitops/gitops"
)

func GetAuth(gitopsObj *gitops.Gitops) (transport.AuthMethod, error) {
	var auth transport.AuthMethod = nil
	var err error = nil

	if gitopsObj.HttpAuth != nil {
		auth = &gitHttp.BasicAuth{
			Username: gitopsObj.HttpAuth.Username,
			Password: gitopsObj.HttpAuth.Password,
		}
	}

	if gitopsObj.CertKey != nil {
		b64decoded, _ := base64.StdEncoding.DecodeString(gitopsObj.CertKey.PrivateKey)
		auth, err = ssh.NewPublicKeys(ssh.DefaultUsername, b64decoded, gitopsObj.CertKey.PrivateKeyPassword)
	}

	return auth, err
}

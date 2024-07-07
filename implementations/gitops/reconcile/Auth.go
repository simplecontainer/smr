package reconcile

import (
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/simplecontainer/smr/implementations/gitops/gitops"
)

func GetAuth(gitopsObj *gitops.Gitops) (transport.AuthMethod, error) {
	var auth transport.AuthMethod
	var err error

	if gitopsObj.HttpAuth != nil {
		auth = &gitHttp.BasicAuth{
			Username: gitopsObj.HttpAuth.Username,
			Password: gitopsObj.HttpAuth.Password,
		}
	}

	if gitopsObj.CertKey != nil {
		auth, err = ssh.NewPublicKeys(ssh.DefaultUsername, []byte(gitopsObj.CertKey.PrivateKey), gitopsObj.CertKey.PrivateKeyPassword)
		return nil, err
	}

	return auth, nil
}

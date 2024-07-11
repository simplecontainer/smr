package gitops

import (
	"github.com/simplecontainer/smr/implementations/gitops/certkey"
	"github.com/simplecontainer/smr/implementations/gitops/httpauth"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/objects"
	"net/http"
)

func (gitops *Gitops) Prepare(client *http.Client) {
	format := f.New("httpauth", gitops.HttpAuthRef.Group, gitops.HttpAuthRef.Identifier, "object")

	var httpAuth v1.HttpAuth
	obj := objects.Object{}
	obj.FindAndConvert(client, format, httpAuth)

	gitops.HttpAuth = &httpauth.HttpAuth{
		Username: httpAuth.Spec.Username,
		Password: httpAuth.Spec.Password,
	}

	var certKey v1.CertKey
	format = f.New("certkey", gitops.CertKeyRef.Group, gitops.CertKeyRef.Identifier, "object")
	obj.FindAndConvert(client, format, certKey)

	gitops.CertKey = &certkey.CertKey{
		Certificate: certKey.Spec.Certificate,
		PublicKey:   certKey.Spec.PublicKey,
		PrivateKey:  certKey.Spec.PrivateKey,
	}
}

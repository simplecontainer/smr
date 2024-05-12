package gitops

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/qdnqn/smr/pkg/certkey"
	"github.com/qdnqn/smr/pkg/database"
	"github.com/qdnqn/smr/pkg/definitions/v1"
	"github.com/qdnqn/smr/pkg/httpauth"
	"github.com/qdnqn/smr/pkg/objects"
)

func (gitops *Gitops) Prepare(db *badger.DB) {
	format := database.Format("httpauth", gitops.HttpAuthRef.Group, gitops.HttpAuthRef.Identifier, "object")

	var httpAuth v1.HttpAuth
	obj := objects.Object{}
	obj.FindAndConvert(db, format, httpAuth)

	gitops.HttpAuth = &httpauth.HttpAuth{
		Username: httpAuth.Spec.Username,
		Password: httpAuth.Spec.Password,
	}

	var certKey v1.CertKey
	format = database.Format("certkey", gitops.CertKeyRef.Group, gitops.CertKeyRef.Identifier, "object")
	obj.FindAndConvert(db, format, certKey)

	gitops.CertKey = &certkey.CertKey{
		Certificate: certKey.Spec.Certificate,
		PublicKey:   certKey.Spec.PublicKey,
		PrivateKey:  certKey.Spec.PrivateKey,
	}
}

package gitops

import (
	"github.com/dgraph-io/badger/v4"
	"smr/pkg/certkey"
	"smr/pkg/database"
	"smr/pkg/definitions"
	"smr/pkg/httpauth"
	"smr/pkg/objects"
)

func (gitops *Gitops) Prepare(db *badger.DB) {
	format := database.Format("httpauth", gitops.HttpAuthRef.Group, gitops.HttpAuthRef.Identifier, "object")

	var httpAuth definitions.HttpAuth
	obj := objects.Object{}
	obj.FindAndConvert(db, format, httpAuth)

	gitops.HttpAuth = &httpauth.HttpAuth{
		Username: httpAuth.Spec.Username,
		Password: httpAuth.Spec.Password,
	}

	var certKey definitions.CertKey
	format = database.Format("certkey", gitops.CertKeyRef.Group, gitops.CertKeyRef.Identifier, "object")
	obj.FindAndConvert(db, format, certKey)

	gitops.CertKey = &certkey.CertKey{
		Certificate: certKey.Spec.Certificate,
		PublicKey:   certKey.Spec.PublicKey,
		PrivateKey:  certKey.Spec.PrivateKey,
	}
}

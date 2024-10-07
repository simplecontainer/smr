package gitops

import (
	"encoding/json"
	"github.com/simplecontainer/smr/implementations/gitops/certkey"
	"github.com/simplecontainer/smr/implementations/gitops/httpauth"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/objects"
)

func (gitops *Gitops) Prepare(client *client.Http, user *authentication.User) (*AuthType, error) {
	if gitops.HttpAuthRef.Group != "" && gitops.HttpAuthRef.Name != "" {
		format := f.New("httpauth", gitops.HttpAuthRef.Group, gitops.HttpAuthRef.Name, "object")

		var httpAuth v1.HttpAuthDefinition
		obj := objects.New(client.Get(user.Username), user)
		err := obj.Find(format)

		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(obj.GetDefinitionByte(), &httpAuth)

		if err != nil {
			return nil, err
		}

		gitops.HttpAuth = &httpauth.HttpAuth{
			Username: httpAuth.Spec.Username,
			Password: httpAuth.Spec.Password,
		}

		return &AuthType{AuthType: httpAuthType}, nil
	}

	if gitops.CertKeyRef.Group != "" && gitops.CertKeyRef.Name != "" {
		var certKey v1.CertKeyDefinition
		format := f.New("certkey", gitops.CertKeyRef.Group, gitops.CertKeyRef.Name, "object")
		obj := objects.New(client.Get(user.Username), user)
		err := obj.Find(format)

		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(obj.GetDefinitionByte(), &certKey)

		if err != nil {
			return nil, err
		}

		gitops.CertKey = &certkey.CertKey{
			Certificate: certKey.Spec.Certificate,
			PublicKey:   certKey.Spec.PublicKey,
			PrivateKey:  certKey.Spec.PrivateKey,
		}

		return &AuthType{AuthType: sshAuthType}, nil
	}

	return nil, nil
}

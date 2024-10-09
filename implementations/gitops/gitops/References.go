package gitops

import (
	"encoding/json"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/objects"
)

func (gitops *Gitops) Prepare(client *client.Http, user *authentication.User) (*AuthType, error) {
	if gitops.Auth.HttpAuthRef.Group != "" && gitops.Auth.HttpAuthRef.Name != "" {
		format := f.New("httpauth", gitops.Auth.HttpAuthRef.Group, gitops.Auth.HttpAuthRef.Name, "object")

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

		gitops.AuthInternal.HttpAuth = &HttpAuth{
			Username: httpAuth.Spec.Username,
			Password: httpAuth.Spec.Password,
		}

		return &AuthType{AuthType: HTTP_AUTH_TYPE}, nil
	}

	if gitops.Auth.CertKeyRef.Group != "" && gitops.Auth.CertKeyRef.Name != "" {
		var certKey v1.CertKeyDefinition
		format := f.New("certkey", gitops.Auth.CertKeyRef.Group, gitops.Auth.CertKeyRef.Name, "object")
		obj := objects.New(client.Get(user.Username), user)
		err := obj.Find(format)

		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(obj.GetDefinitionByte(), &certKey)

		if err != nil {
			return nil, err
		}

		gitops.AuthInternal.CertKey = &CertKey{
			Certificate: certKey.Spec.Certificate,
			PublicKey:   certKey.Spec.PublicKey,
			PrivateKey:  certKey.Spec.PrivateKey,
		}

		return &AuthType{AuthType: SSH_AUTH_TYPE}, nil
	}

	return nil, nil
}

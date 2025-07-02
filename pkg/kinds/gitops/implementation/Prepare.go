package implementation

import (
	"errors"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
)

func (gitops *Gitops) Prepare(client *clients.Http, user *authentication.User) error {
	obj := objects.New(client.Get(user.Username), user)
	references, err := gitops.GetDefinition().ResolveReferences(obj)

	if err != nil {
		return err
	}

	for _, reference := range references {
		switch reference.GetKind() {
		case static.KIND_HTTPAUTH:
			return gitops.Gitops.Git.Auth.Http(reference.(*v1.HttpAuthDefinition))
		case static.KIND_CERTKEY:
			return gitops.Gitops.Git.Auth.Ssh(reference.(*v1.CertKeyDefinition))
		default:
			return errors.New("reference kind is not implemented for this type of object")
		}
	}

	return nil
}

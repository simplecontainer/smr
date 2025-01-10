package implementation

import (
	"errors"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
)

func (gitops *Gitops) Prepare(client *client.Http, user *authentication.User) error {
	obj := objects.New(client.Get(user.Username), user)
	references, err := gitops.Definition.ResolveReferences(obj)

	if err != nil {
		return err
	}

	var auth transport.AuthMethod

	for _, reference := range references {
		switch reference.GetKind() {
		case static.KIND_HTTPAUTH:
			auth, err = gitops.GenerateHttpAuth(reference.(*v1.HttpAuthDefinition))

			if err != nil {
				return err
			}

			gitops.AuthResolved = auth
			break
		case static.KIND_CERTKEY:
			auth, err = gitops.GenerateSshAuth(reference.(*v1.CertKeyDefinition))

			if err != nil {
				return err
			}

			gitops.AuthResolved = auth
			break
		default:
			return errors.New("reference kind is not implemented for this type of object")
		}
	}

	return nil
}

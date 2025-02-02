package template

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/smaps"
	"github.com/simplecontainer/smr/pkg/static"
)

func Lookup(placeholder string, client *client.Http, user *authentication.User, runtime *smaps.Smap, dependencies []f.Format) (string, error) {
	format := f.NewFromString(placeholder)

	if !format.Compliant() {
		format = format.Inverse().(f.Format)
	}

	// Handle case when format is specified in kind compliant format
	// eg:
	//	simplecontainer.io/v1/kind/configuration/group/name:element
	//  simplecontainer.io/v1/kind/secret/group/name:element
	//  configuration/group/name:element (missing prefix)
	//  secret/group/name:element (missing prefix)
	//  runtime/container/configuration:element (not kind but container runtime)

	switch format.GetKind() {
	case "secret":
		obj := objects.New(client.Clients[user.Username], user)
		name, key, err := Extract(format.GetName())

		if err != nil {
			return placeholder, err
		}

		var fc f.Format
		if format.Compliant() {
			fc = f.New(format.GetPrefix(), format.GetVersion(), format.GetCategory(), format.GetKind(), format.GetGroup(), name)
		} else {
			fc = f.New(static.SMR_PREFIX, "kind", format.GetKind(), format.GetGroup(), name)
		}

		obj.Find(fc)

		if !obj.Exists() {
			return placeholder, errors.New(fmt.Sprintf("lookup: object doesn't exists %s", fc.ToString()))
		}

		secret := v1.SecretDefinition{}

		err = json.Unmarshal(obj.GetDefinitionByte(), &secret)

		if err != nil {
			return placeholder, err
		}

		_, ok := secret.Spec.Data[key]

		if !ok {
			return placeholder, errors.New(
				fmt.Sprintf("missing field in the configuration resource: %s", placeholder),
			)
		}

		decoded, err := base64.StdEncoding.DecodeString(secret.Spec.Data[key])

		if err != nil {
			return placeholder, err
		}

		return string(decoded), nil
	case "configuration":
		obj := objects.New(client.Clients[user.Username], user)
		name, key, err := Extract(format.GetName())

		if err != nil {
			return placeholder, err
		}

		var fc f.Format
		if format.Compliant() {
			fc = f.New(format.GetPrefix(), format.GetVersion(), format.GetCategory(), format.GetKind(), format.GetGroup(), name)
		} else {
			fc = f.New(static.SMR_PREFIX, "kind", format.GetKind(), format.GetGroup(), name)
		}

		obj.Find(fc)

		if !obj.Exists() {
			return placeholder, errors.New(fmt.Sprintf("lookup: object doesn't exists %s", fc.ToString()))
		}

		dependencies = append(dependencies, fc)

		configuration := v1.ConfigurationDefinition{}

		err = json.Unmarshal(obj.GetDefinitionByte(), &configuration)

		if err != nil {
			return placeholder, err
		}

		_, ok := configuration.Spec.Data[key]

		if !ok {
			return placeholder, errors.New(
				fmt.Sprintf("missing field in the configuration resource: %s", placeholder),
			)
		}

		return configuration.Spec.Data[key], nil
	case "runtime":
		// Handle case when format is specified in kind non-compliant format
		// eg:
		//	runtime/container/configuration:element

		name, key, err := Extract(format.GetName())

		if err != nil {
			return placeholder, err
		}

		switch name {
		case "configuration":
			var value any
			var ok = false

			if runtime == nil {
				ok = false
			} else {
				value, ok = runtime.Map.Load(key)
			}

			if ok {
				return value.(string), nil
			} else {
				return placeholder, errors.New(fmt.Sprintf("container runtime configuration is missing: %s", placeholder))
			}
		default:
			return placeholder, errors.New(fmt.Sprintf("unsupported lookup: %s", placeholder))
		}
	default:
		return placeholder, errors.New(fmt.Sprintf("unsupported lookup: %s", placeholder))
	}
}

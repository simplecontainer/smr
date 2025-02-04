package template

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts"
	v1 "github.com/simplecontainer/smr/pkg/definitions/v1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/smaps"
	"github.com/simplecontainer/smr/pkg/static"
)

func Lookup(placeholder string, client *client.Http, user *authentication.User, runtime *smaps.Smap, dependencies []f.Format, depth int) (string, error) {
	if depth > 1 {
		return placeholder, errors.New("depth is too big consider restructuring definition files")
	}

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
		name, key, err := Extract(format.GetName())

		if err != nil {
			return placeholder, err
		}

		bytes, _, err := Fetch(format, name, client, user)

		if err != nil {
			return placeholder, err
		}

		secret := v1.SecretDefinition{}
		err = json.Unmarshal(bytes, &secret)

		if err != nil {
			return placeholder, err
		}

		_, ok := secret.Spec.Data[key]

		if !ok {
			return placeholder, errors.New(
				fmt.Sprintf("missing field in the secret resource: %s", placeholder),
			)
		}

		return secret.Spec.Data[key], nil
	case "configuration":
		name, key, err := Extract(format.GetName())

		if err != nil {
			return placeholder, err
		}

		bytes, formatClean, err := Fetch(format, name, client, user)

		if err != nil {
			return placeholder, err
		}

		dependencies = append(dependencies, formatClean)

		configuration := v1.ConfigurationDefinition{}

		err = json.Unmarshal(bytes, &configuration)

		if err != nil {
			return placeholder, err
		}

		_, ok := configuration.Spec.Data[key]

		if !ok {
			return placeholder, errors.New(
				fmt.Sprintf("missing field in the configuration resource: %s", placeholder),
			)
		}

		// Since configuration can also have templated values parse it. eg container config-> configuration-> secret
		parsed, _, err := Parse(placeholder, configuration.Spec.Data[key], client, user, runtime, depth+1)

		return parsed, nil
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

func Fetch(format contracts.Format, name string, client *client.Http, user *authentication.User) ([]byte, f.Format, error) {
	obj := objects.New(client.Clients[user.Username], user)

	var formatNoKey f.Format
	if format.Compliant() {
		formatNoKey = f.New(format.GetPrefix(), format.GetVersion(), format.GetCategory(), format.GetKind(), format.GetGroup(), name)
	} else {
		formatNoKey = f.New(static.SMR_PREFIX, "kind", format.GetKind(), format.GetGroup(), name)
	}

	obj.Find(formatNoKey)

	if !obj.Exists() {
		return nil, formatNoKey, errors.New(fmt.Sprintf("lookup: object doesn't exists %s", formatNoKey.ToString()))
	}

	return obj.GetDefinitionByte(), formatNoKey, nil
}

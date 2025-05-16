package template

import (
	"encoding/base64"
	"errors"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/smaps"
	"strings"
	"text/template"
)

func Parse(name string, value string, client *clients.Http, user *authentication.User, runtime *smaps.Smap, depth int) (string, []f.Format, error) {
	var dependencies = make([]f.Format, 0)

	variables := Variables{Values: make(map[string]interface{})}

	if runtime != nil {
		runtime.Map.Range(func(key any, value any) bool {
			variables.Values[key.(string)] = value.(string)
			return true
		})
	}

	t := New(name, value, variables, template.FuncMap{
		"fqdn": func(name string) (string, error) {
			return FQDN(name), nil
		},
		"lookup": func(placeholder string) (string, error) {
			return Lookup(placeholder, client, user, runtime, dependencies, depth)
		},
		"base64decode": func(input string) (string, error) {
			decoded, err := base64.StdEncoding.DecodeString(input)

			if err != nil {
				return input, err
			}

			return string(decoded), nil
		},
		"base64encode": func(input string) string {
			return base64.StdEncoding.EncodeToString([]byte(input))
		},
	})

	parsed, err := t.Parse("((", "))")

	if err != nil {
		return value, dependencies, err
	}

	return parsed, dependencies, nil
}

func Extract(namekey string) (string, string, error) {
	split := strings.Split(namekey, ":")

	if len(split) == 2 {
		return split[0], split[1], nil
	}

	return "", "", errors.New("invalid format for name:key")
}

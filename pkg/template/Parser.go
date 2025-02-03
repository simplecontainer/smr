package template

import (
	"errors"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/smaps"
	"html/template"
	"strings"
)

func Parse(name string, value string, client *client.Http, user *authentication.User, runtime *smaps.Smap, depth int) (string, []f.Format, error) {
	var dependencies = make([]f.Format, 0)

	variables := Variables{Values: make(map[string]interface{})}

	if runtime != nil {
		runtime.Map.Range(func(key any, value any) bool {
			variables.Values[key.(string)] = value.(string)
			return true
		})
	}

	t := New(name, value, variables, template.FuncMap{
		"lookup": func(placeholder string) (string, error) {
			return Lookup(placeholder, client, user, runtime, dependencies, depth)
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

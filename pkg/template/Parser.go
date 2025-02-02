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

func Parse(name string, value string, client *client.Http, user *authentication.User, runtime *smaps.Smap) (string, []f.Format, error) {
	var dependencies = make([]f.Format, 0)

	t := New(name, value, Variables{}, template.FuncMap{
		"lookup": func(placeholder string) (string, error) {
			return Lookup(placeholder, client, user, runtime, dependencies)
		},
	})

	parsed, err := t.Parse()

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

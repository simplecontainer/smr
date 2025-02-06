package objects

import (
	"github.com/go-playground/assert/v2"
	"github.com/r3labs/diff/v3"
	"github.com/wI2L/jsondiff"

	"testing"
)

func TestDiff(t *testing.T) {
	obj1 := []byte(`{"kind":"containers","prefix":"simplecontainer.io/v1","meta":{"group":"traefik","name":"traefik","labels":{"test":"testing"},"runtime":{"owner":{"Kind":"","Group":"","Name":""},"node":2,"nodeName":"smr-agent-2"}},"spec":{"image":"traefik","tag":"v2.5","dependencies":[{"prefix":"simplecontainer.io/v1","name":"*","group":"mysql","timeout":"30s"}],"ports":[{"container":"80","host":"80"},{"container":"443","host":"443"},{"container":"8080","host":"8888"}],"volumes":[{"name":"","type":"bind","hostPath":"/var/run/docker.sock","mountPoint":"/var/run/docker.sock"}],"resources":[{"Name":"config","Group":"traefik","Key":"traefik-configuration","MountPoint":"/etc/traefik/traefik.yml"}],"replicas":1,"spread":{"spread":""}},"state":{"Options":[]}}
`)
	obj2 := []byte(`{"kind":"containers","prefix":"simplecontainer.io/v1","meta":{"group":"traefik","name":"traefik","labels":{"test":"testing"},"runtime":{"owner":{"Kind":"","Group":"","Name":""},"node":2,"nodeName":"smr-agent-2"}},"spec":{"image":"traefik","tag":"v2.5","dependencies":[{"prefix":"simplecontainer.io/v1","name":"*","group":"mysql","timeout":"30s"}],"ports":[{"container":"80","host":"80"},{"container":"443","host":"443"},{"container":"8080","host":"8888"}],"volumes":[{"name":"","type":"bind","hostPath":"/var/run/docker.sock","mountPoint":"/var/run/docker.sock"}],"resources":[{"Name":"config","Group":"traefik","Key":"traefik-configuration","MountPoint":"/etc/traefik/traefik.yml"}],"replicas":2,"spread":{"spread":""}},"state":{"Options":[]}}
`)

	type Wanted struct {
		changelog      jsondiff.Patch
		changelogEmpty jsondiff.Patch
		err            error
		errEmpty       error
		diff           diff.Change
	}

	type Parameters struct {
		empty []byte
		obj1  []byte
		obj2  []byte
	}

	changelog, err := jsondiff.CompareJSON(obj1, obj2)
	changelogEmpty, errEmpty := jsondiff.CompareJSON([]byte(`{}`), obj2)

	testCases := []struct {
		name       string
		mockFunc   func()
		wanted     Wanted
		parameters Parameters
	}{
		{
			"Valid JSON",
			func() {
			},
			Wanted{
				changelog:      changelog,
				changelogEmpty: changelogEmpty,
				err:            err,
				errEmpty:       errEmpty,
				diff:           diff.Change{},
			},
			Parameters{
				empty: []byte(`{}`),
				obj1:  obj1,
				obj2:  obj2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()

			changelogT, errT := jsondiff.CompareJSON(tc.parameters.obj1, tc.parameters.obj2)
			changelogEmptyT, errEmptyT := jsondiff.CompareJSON(tc.parameters.empty, tc.parameters.obj2)

			assert.Equal(t, tc.wanted.changelog[0].Path, changelogT[0].Path)
			assert.Equal(t, tc.wanted.changelogEmpty[0].Path, changelogEmptyT[0].Path)
			assert.Equal(t, tc.wanted.changelog, changelogT)
			assert.Equal(t, tc.wanted.err, errT)
			assert.Equal(t, tc.wanted.errEmpty, errEmptyT)
		})
	}
}

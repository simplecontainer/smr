package startup

import (
	"bytes"
	"fmt"
	"github.com/go-playground/assert/v2"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/static"
	"io"
	"os"
	"testing"
)

func TestGetEnvironmentInfo(t *testing.T) {
	type Wanted struct {
		environment *configuration.Environment
	}

	type Parameters struct {
	}

	testCases := []struct {
		name       string
		mockFunc   func()
		wanted     Wanted
		parameters Parameters
	}{
		{
			"Valid configuration",
			func() {
			},
			Wanted{
				environment: &configuration.Environment{
					HOMEDIR:    "",
					OPTDIR:     "/opt/smr",
					PROJECTDIR: "",
					PROJECT:    static.PROJECT,
					AGENTIP:    "",
				},
			},
			Parameters{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()

			HOMEDIR, err := os.UserHomeDir()
			if err != nil {
				panic(err.Error())
			}

			tc.wanted.environment.HOMEDIR = HOMEDIR
			tc.wanted.environment.PROJECTDIR = fmt.Sprintf("%s/%s/%s", HOMEDIR, static.ROOTDIR, tc.wanted.environment.PROJECT)
			tc.wanted.environment.AGENTIP = GetOutboundIP().String()

			environment := GetEnvironmentInfo()

			assert.Equal(t, tc.wanted.environment, environment)
		})
	}
}

func TestLoad(t *testing.T) {
	const validConfiguration = `
target: "development"
root: "/home/qdnqn/smr/smr"
`

	type Wanted struct {
		configuration *configuration.Configuration
		error         error
	}

	type Parameters struct {
		configuration *configuration.Configuration
		reader        *io.Reader
	}

	testCases := []struct {
		name       string
		mockFunc   func()
		wanted     Wanted
		parameters Parameters
	}{
		{
			"Valid configuration",
			func() {
			},
			Wanted{
				configuration: &configuration.Configuration{
					Target:      "development",
					Root:        "/home/qdnqn/smr/smr",
					Environment: nil,
					Flags:       configuration.Flags{},
				},
				error: nil,
			},
			Parameters{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()

			configObj, err := Load(bytes.NewBuffer([]byte(validConfiguration)))

			assert.Equal(t, tc.wanted.configuration, configObj)
			assert.Equal(t, tc.wanted.error, err)
		})
	}
}

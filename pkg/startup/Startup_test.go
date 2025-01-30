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
					Home:          "",
					NodeIP:        "",
					NodeDirectory: "",
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

			tc.wanted.environment.Home = HOMEDIR
			tc.wanted.environment.NodeDirectory = fmt.Sprintf("%s/%s/%s", HOMEDIR, static.ROOTDIR, tc.wanted.environment.PROJECT)
			tc.wanted.environment.AGENTIP = GetOutboundIP().String()

			environment := GetEnvironmentInfo()

			assert.Equal(t, tc.wanted.environment, environment)
		})
	}
}

func TestLoad(t *testing.T) {
	const validConfiguration = `
target: development
root: /home/smr-agent/smr/smr
optroot: /opt/smr
domain: localhost
externalIP: 127.0.0.1
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
					Root:        "/home/smr-agent/smr/smr",
					ExternalIP:  "127.0.0.1",
					Domain:      "localhost",
					Flags:       configuration.Flags{},
					Environment: GetEnvironmentInfo(),
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

			assert.Equal(t, tc.wanted.configuration.Environment, configObj.Environment)
			assert.Equal(t, tc.wanted.configuration.Target, configObj.Target)
			assert.Equal(t, tc.wanted.configuration.Root, configObj.Root)
			assert.Equal(t, tc.wanted.configuration.ExternalIP, configObj.ExternalIP)
			assert.Equal(t, tc.wanted.configuration.Domain, configObj.Domain)
			assert.Equal(t, tc.wanted.error, err)
		})
	}
}

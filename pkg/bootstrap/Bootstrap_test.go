package bootstrap

import (
	"fmt"
	"github.com/go-playground/assert/v2"
	"github.com/simplecontainer/smr/pkg/configuration"
	"github.com/simplecontainer/smr/pkg/startup"
	"os"
	"testing"
)

func TestCreateProject(t *testing.T) {
	type Wanted struct {
		created []string
		error   error
	}

	type Parameters struct {
		project       string
		configuration *configuration.Configuration
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
				created: []string{
					fmt.Sprintf("%s/smr/test/config", os.Getenv("HOME")),
					fmt.Sprintf("%s/smr/test/persistent", os.Getenv("HOME")),
					fmt.Sprintf("%s/smr/test/persistent/smr", os.Getenv("HOME")),
				},
				error: nil,
			},
			Parameters{
				project:       "test",
				configuration: configuration.NewConfig(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()

			tc.parameters.configuration.Environment = startup.GetEnvironmentInfo()

			created, err := CreateProject(tc.parameters.project, tc.parameters.configuration)

			assert.Equal(t, tc.wanted.created, created)
			assert.Equal(t, tc.wanted.error, err)

			// cleanup and test also
			err = DeleteProject(tc.parameters.project, tc.parameters.configuration)
			assert.Equal(t, tc.wanted.error, err)
		})
	}
}

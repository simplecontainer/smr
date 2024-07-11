package template

import (
	"github.com/go-playground/assert/v2"
	"github.com/golang/mock/gomock"
	"github.com/simplecontainer/smr/pkg/f"
	mock_objects "github.com/simplecontainer/smr/pkg/objects/mock"
	"testing"
)

func TestParseTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)

	const example1 = `
    {
	  "meta":{
		"group":"mysql",
		"identifier":"*"
	  },
	  "spec":{
		"data":{
		  "password":"{{ secret.mysql.mysql.password }}"
		}
	  }
	}
	`

	const example2 = `
    {
	  "meta":{
		"group":"mysql",
		"identifier":"mysql"
	  },
	  "spec":{
		"data":{
		  "password":"{{ secret.mysql.mysql.password }}"
		}
	  }
	}
	`

	objMock := mock_objects.NewMockObjectInterface(ctrl)

	type Wanted struct {
		parsed       map[string]string
		dependencies []*f.Format
		parameters   map[string]string
		error        error
	}

	type Parameters struct {
		values map[string]string
		format *f.Format
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
				objMock.EXPECT().Find(f.NewFromString("configuration.mysql.*.object")).Return(nil).Times(1)
				objMock.EXPECT().Find(f.NewFromString("configuration.mysql.mysql.object")).Return(nil).Times(1)
				objMock.EXPECT().Exists().Return(true).Times(2)
				objMock.EXPECT().GetDefinitionByte().Return(
					[]byte(example1),
				).Times(1)
				objMock.EXPECT().GetDefinitionByte().Return(
					[]byte(example2),
				).Times(1)
				objMock.EXPECT().Add(f.NewFromString("configuration.mysql.test-test-1.username"), "root").Return(nil).Times(1)
			},
			Wanted{
				parsed: map[string]string{
					"password":  "{{ secret.mysql.mysql.password }}",
					"password2": "{{ secret.mysql.mysql.password }}",
				},
				dependencies: []*f.Format{
					f.NewFromString("configuration.mysql.*.object"),
					f.NewFromString("configuration.mysql.mysql.object"),
				},

				error: nil,
			},
			Parameters{
				values: map[string]string{
					"password":  "{{ configuration.mysql.*.password }}",
					"password2": "{{ configuration.mysql.mysql.password }}",
					"username":  "root",
				},
				format: f.NewFromString("configuration.mysql.test-test-1"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()

			parsed, dependencies, err := ParseTemplate(objMock, tc.parameters.values, tc.parameters.format)

			assert.Equal(t, tc.wanted.parsed, parsed)
			assert.Equal(t, tc.wanted.dependencies, dependencies)
			assert.Equal(t, tc.wanted.error, err)
		})
	}
}

func TestParseSecretTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)

	objMock := mock_objects.NewMockObjectInterface(ctrl)

	type Wanted struct {
		parsed string
		error  error
	}

	type Parameters struct {
		value string
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
				objMock.EXPECT().Find(f.NewFromString("secret.mysql.mysql.password")).Return(nil).Times(1)
				objMock.EXPECT().Exists().Return(true).Times(1)
				objMock.EXPECT().GetDefinitionString().Return("123456").Times(1)
			},
			Wanted{
				parsed: "123456",
				error:  nil,
			},
			Parameters{
				value: "{{ secret.mysql.mysql.password }}",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc()

			parsed, err := ParseSecretTemplate(objMock, tc.parameters.value)

			assert.Equal(t, tc.wanted.parsed, parsed)
			assert.Equal(t, tc.wanted.error, err)
		})
	}
}

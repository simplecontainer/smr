package template

import (
	"fmt"
	"github.com/go-playground/assert/v2"
	"github.com/golang/mock/gomock"
	"github.com/simplecontainer/smr/pkg/f"
	mock_objects "github.com/simplecontainer/smr/pkg/objects/mock"
	"testing"
)

func TestParseTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)

	objMock := mock_objects.NewMockObjectInterface(ctrl)

	objMock.EXPECT().Find(f.NewFromString("configuration.mysql.*.object")).Return(nil).Times(1)
	objMock.EXPECT().Find(f.NewFromString("configuration.mysql.mysql.object")).Return(nil).Times(1)
	objMock.EXPECT().Exists().Return(true).Times(2)
	objMock.EXPECT().GetDefinitionByte().Return(
		[]byte("{ \"meta\": { \"group\": \"mysql\", \"identifier\": \"*\" }, \"spec\": { \"data\": { \"password\": \"{{ secret.mysql.mysql.password }}\" } } }"),
	).Times(1)
	objMock.EXPECT().GetDefinitionByte().Return(
		[]byte("{ \"meta\": { \"group\": \"mysql\", \"identifier\": \"mysql\" }, \"spec\": { \"data\": { \"password\": \"{{ secret.mysql.mysql.password }}\" } } }"),
	).Times(1)
	objMock.EXPECT().Add(f.NewFromString("configuration.mysql.test-test-1.username"), "root").Return(nil).Times(1)

	wantedParsed := map[string]string{
		"password":  "{{ secret.mysql.mysql.password }}",
		"password2": "{{ secret.mysql.mysql.password }}",
	}

	wantedDependency := []*f.Format{}
	wantedDependency = append(wantedDependency, f.NewFromString("configuration.mysql.*.object"))
	wantedDependency = append(wantedDependency, f.NewFromString("configuration.mysql.mysql.object"))

	parsedMap, dependencyList, err := ParseTemplate(
		objMock,
		map[string]string{
			"password":  "{{ configuration.mysql.*.password }}",
			"password2": "{{ configuration.mysql.mysql.password }}",
			"username":  "root",
		},
		f.NewFromString("configuration.mysql.test-test-1"),
	)

	if err != nil {
		fmt.Println(err.Error())
	}

	assert.Equal(t, wantedParsed, parsedMap)
	assert.Equal(t, wantedDependency, dependencyList)
	assert.Equal(t, nil, err)
}

func TestParseSecretTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)

	objMock := mock_objects.NewMockObjectInterface(ctrl)

	objMock.EXPECT().Find(f.NewFromString("secret.mysql.mysql.password")).Return(nil).Times(1)
	objMock.EXPECT().Exists().Return(true).Times(1)
	objMock.EXPECT().GetDefinitionString().Return("123456").Times(1)

	wantedParsed := "123456"
	parsedSecret, err := ParseSecretTemplate(
		objMock,
		"{{ secret.mysql.mysql.password }}",
	)

	if err != nil {
		fmt.Println(err.Error())
	}

	assert.Equal(t, wantedParsed, parsedSecret)
	assert.Equal(t, nil, err)
}

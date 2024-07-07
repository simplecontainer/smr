package objects

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
	"net/http"
	"reflect"
	"strings"
	"time"
)

func New() *Object {
	return &Object{
		Changelog:        diff.Changelog{},
		definition:       map[string]any{},
		definitionString: "",
		definitionByte:   make([]byte, 0),
		exists:           false,
		changed:          false,
		created:          time.Now(),
		updated:          time.Now(),
	}
}

func (obj *Object) GetDefinitionString() string {
	return obj.definitionString
}

func (obj *Object) GetDefinition() map[string]any {
	return obj.definition
}

func (obj *Object) GetDefinitionByte() []byte {
	return obj.definitionByte
}

func (obj *Object) Add(client *http.Client, format FormatStructure, data string) error {
	URL := fmt.Sprintf("https://smr-agent.docker.private:1443/api/v1/database/create/%s", format.ToString())
	response := SendRequest(client, URL, "POST", map[string]string{"value": data})

	logger.Log.Debug("object add", zap.String("URL", URL), zap.String("data", data))

	if response.Success {
		return nil
	} else {
		return errors.New(response.ErrorExplanation)
	}
}

func (obj *Object) Update(client *http.Client, format FormatStructure, data string) error {
	URL := fmt.Sprintf("https://smr-agent.docker.private:1443/api/v1/database/update/%s", format.ToString())
	response := SendRequest(client, URL, "PUT", map[string]string{"value": data})

	logger.Log.Debug("object update", zap.String("URL", URL), zap.String("data", data))

	if response.Success {
		return nil
	} else {
		return errors.New(response.ErrorExplanation)
	}
}

func (obj *Object) Find(client *http.Client, format FormatStructure) error {
	URL := fmt.Sprintf("https://smr-agent.docker.private:1443/api/v1/database/get/%s", format.ToString())
	response := SendRequest(client, URL, "GET", nil)

	logger.Log.Debug("object find", zap.String("URL", URL))

	if response.Success {
		for key, value := range response.Data {
			if value == nil {
				continue
			}

			if strings.Contains(key, "object") {
				b64decoded, err := base64.StdEncoding.DecodeString(value.(string))

				data := make(map[string]any)
				err = json.Unmarshal(b64decoded, &data)

				if err != nil {
					return err
				}

				obj.definition = data
				obj.definitionByte = b64decoded
			} else {
				b64decoded, _ := base64.StdEncoding.DecodeString(value.(string))
				obj.definitionString = string(b64decoded)
			}
		}

	} else {
		return errors.New(response.ErrorExplanation)
	}

	obj.changed = false
	obj.exists = true

	return nil
}

func FindMany(client *http.Client, format FormatStructure) (map[string]*Object, error) {
	var objects = make(map[string]*Object)

	URL := fmt.Sprintf("https://smr-agent.docker.private:1443/api/v1/database/keys/prefix/%s", format.ToString())
	response := SendRequest(client, URL, "GET", nil)

	logger.Log.Debug("object find many", zap.String("URL", URL))

	if response.Success {
		for key, value := range response.Data {
			if strings.Contains(key, "object") {
				fmt.Println(value.(string))

				b64decoded, err := base64.StdEncoding.DecodeString(value.(string))

				if err != nil {
					return nil, err
				}

				fmt.Println(string(b64decoded))

				data := make(map[string]any)
				err = json.Unmarshal(b64decoded, &data)

				if err != nil {
					return nil, err
				}

				obj := New()
				obj.definition = data
				obj.definitionByte = b64decoded

				objects[key] = obj
			} else {
				b64decoded, _ := base64.StdEncoding.DecodeString(value.(string))

				obj := New()
				obj.definitionString = string(b64decoded)
				obj.definitionByte = b64decoded

				objects[key] = obj
			}
		}
	} else {
		return nil, errors.New(response.ErrorExplanation)
	}

	return objects, nil
}

func (obj *Object) Remove(client *http.Client, format FormatStructure) (bool, error) {
	URL := fmt.Sprintf("https://smr-agent.docker.private:1443/api/v1/database/keys/%s", format.ToString())
	response := SendRequest(client, URL, "DELETE", nil)

	logger.Log.Debug("object remove", zap.String("URL", URL))

	if response.Success {
		return true, nil
	} else {
		return false, errors.New(response.ErrorExplanation)
	}
}

func (obj *Object) Diff(definition string) bool {
	data := make(map[string]any)
	err := json.Unmarshal([]byte(definition), &data)

	if err != nil {
		return true
	}

	var changelog diff.Changelog

	if reflect.DeepEqual(obj.definition, data) {
		obj.changed = false
	} else {
		changelog, _ = diff.Diff(obj.definition, data)
		obj.Changelog = changelog
		obj.changed = true
	}

	return obj.changed
}

func (obj *Object) Exists() bool {
	return obj.exists
}

func (obj *Object) ChangeDetected() bool {
	return obj.changed
}

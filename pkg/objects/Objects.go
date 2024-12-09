package objects

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
	"reflect"
	"strings"
	"time"
)

//go:generate mockgen -source=Interface.go -destination=mock/Interface.go

func New(client *client.Client, user *authentication.User) *Object {
	return &Object{
		Changelog:        diff.Changelog{},
		client:           client,
		definition:       map[string]any{},
		DefinitionString: "",
		definitionByte:   make([]byte, 0),
		exists:           false,
		changed:          false,
		Created:          time.Now(),
		Updated:          time.Now(),
		User:             user,
	}
}

func (obj *Object) GetDefinitionString() string {
	return obj.DefinitionString
}

func (obj *Object) GetDefinition() map[string]any {
	return obj.definition
}

func (obj *Object) GetDefinitionByte() []byte {
	return obj.definitionByte
}

func (obj *Object) Add(format *f.Format, data string) error {
	URL := fmt.Sprintf("https://%s/api/v1/database/propose/%s", obj.client.API, format.ToString())
	response := SendRequest(obj.client.Http, URL, "POST", []byte(data))

	logger.Log.Debug("object add", zap.String("URL", URL), zap.String("data", data))

	if response.Success {
		URL = fmt.Sprintf("https://%s/api/v1/database/propose/%s.auth", obj.client.API, format.ToString())
		response = SendRequest(obj.client.Http, URL, "POST", obj.User.ToBytes())

		logger.Log.Debug("object auth remove", zap.String("URL", URL))

		if !response.Success {
			return errors.New(response.ErrorExplanation)
		} else {
			return nil
		}
	} else {
		return errors.New(response.ErrorExplanation)
	}
}

func (obj *Object) Update(format *f.Format, data string) error {
	URL := fmt.Sprintf("https://%s/api/v1/database/propose/%s", obj.client.API, format.ToString())
	response := SendRequest(obj.client.Http, URL, "PUT", []byte(data))

	logger.Log.Debug("object update", zap.String("URL", URL), zap.String("data", data))

	if response.Success {
		return nil
	} else {
		return errors.New(response.ErrorExplanation)
	}
}

func (obj *Object) Find(format *f.Format) error {
	URL := fmt.Sprintf("https://%s/api/v1/database/get/%s", obj.client.API, format.ToString())
	response := SendRequest(obj.client.Http, URL, "GET", nil)

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
				obj.DefinitionString = value.(string)
			} else {
				b64decoded, _ := base64.StdEncoding.DecodeString(value.(string))
				obj.DefinitionString = string(b64decoded)
				obj.definitionByte = b64decoded
			}
		}
	} else {
		return errors.New(response.ErrorExplanation)
	}

	obj.changed = false
	obj.exists = true

	return nil
}

func (obj *Object) FindMany(format *f.Format) (map[string]*Object, error) {
	var objects = make(map[string]*Object)

	URL := fmt.Sprintf("https://%s/api/v1/database/keys/%s", obj.client.API, format.ToString())
	response := SendRequest(obj.client.Http, URL, "GET", nil)

	logger.Log.Debug("object find many", zap.String("URL", URL))

	if response.Success {
		if response.Data["keys"] != nil {
			keys := response.Data["keys"].([]interface{})

			for _, value := range keys {
				objTmp := New(obj.client, obj.User)
				err := objTmp.Find(f.NewFromString(value.(string)))

				if err != nil {
					fmt.Println(err)
				}

				if !strings.HasSuffix(value.(string), ".auth") {
					objects[value.(string)] = objTmp
				}
			}
		}
	} else {
		return nil, errors.New(response.ErrorExplanation)
	}

	return objects, nil
}

func (obj *Object) Remove(format *f.Format) (bool, error) {
	prefix := format.ToString()

	if !format.Full() {
		// Append dot to the end of the format so that we delimit what we deleting from the kv-store
		prefix += "."
	}

	URL := fmt.Sprintf("https://%s/api/v1/database/keys/%s", obj.client.API, prefix)
	response := SendRequest(obj.client.Http, URL, "DELETE", nil)

	logger.Log.Debug("object remove", zap.String("URL", URL))

	if response.Success {
		URL = fmt.Sprintf("https://%s/api/v1/database/keys/%s.auth", obj.client.API, prefix)
		response = SendRequest(obj.client.Http, URL, "DELETE", nil)

		logger.Log.Debug("object auth remove", zap.String("URL", URL))

		if !response.Success {
			return false, errors.New(response.ErrorExplanation)
		} else {
			return true, nil
		}
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

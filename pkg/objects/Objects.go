package objects

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
	"reflect"
	"strings"
	"time"
)

//go:generate mockgen -source=Interface.go -destination=mock/Interface.go

func New(client *client.Client, user *authentication.User) *Object {
	return &Object{
		Changelog:  diff.Changelog{},
		client:     client,
		Definition: map[string]any{},
		String:     "",
		Byte:       make([]byte, 0),
		exists:     false,
		changed:    false,
		Created:    time.Now(),
		Updated:    time.Now(),
		User:       user,
	}
}

func (obj *Object) GetDefinitionString() string {
	return obj.String
}

func (obj *Object) GetDefinition() map[string]any {
	return obj.Definition
}

func (obj *Object) GetDefinitionByte() []byte {
	return obj.Byte
}

func (obj *Object) Add(format *f.Format, data string) error {
	URL := fmt.Sprintf("https://%s/api/v1/database/propose/%s/%s", obj.client.API, format.Category, format.ToString())
	response := SendRequest(obj.client.Http, URL, "POST", []byte(data))

	logger.Log.Debug("object add", zap.String("URL", URL), zap.String("data", data))

	if response.Success {
		URL = fmt.Sprintf("https://%s/api/v1/database/propose/%s/%s.auth", obj.client.API, static.CATEGORY_PLAIN, format.ToString())
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
	URL := fmt.Sprintf("https://%s/api/v1/database/propose/%s/%s", obj.client.API, format.Category, format.ToString())
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
		obj.Byte, _ = response.Data.MarshalJSON()
		obj.String = string(obj.Byte)

		err := json.Unmarshal(obj.Byte, &obj.Definition)

		if err != nil {
			logger.Log.Error(err.Error())
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
		if response.Data != nil {
			var keys []string

			bytes, _ := response.Data.MarshalJSON()

			json.Unmarshal(bytes, &keys)

			for _, key := range keys {
				objTmp := New(obj.client, obj.User)
				err := objTmp.Find(f.NewFromString(key))

				if err != nil {
					return objects, err
				}

				if !strings.HasSuffix(key, ".auth") {
					objects[key] = objTmp
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

	if reflect.DeepEqual(obj.Definition, data) {
		obj.changed = false
	} else {
		changelog, _ = diff.Diff(obj.Definition, data)
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

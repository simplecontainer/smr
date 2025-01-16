package objects

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/pkg/acks"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"go.uber.org/zap"
	"net/http"
	"time"
)

//go:generate mockgen -source=Interface.go -destination=mock/Interface.go

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func New(client *client.Client, user *authentication.User) contracts.ObjectInterface {
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

func (obj *Object) Propose(format contracts.Format, data []byte) (uuid.UUID, error) {
	URL := fmt.Sprintf("https://%s/api/v1/database/propose/%s/%s", obj.client.API, format.GetCategory(), format.ToString())
	response := network.Send(obj.client.Http, URL, "POST", data)

	logger.Log.Debug("object add", zap.String("URL", URL), zap.String("data", string(data)))

	if response.Success {
		return format.GetUUID(), nil
	} else {
		return uuid.UUID{}, errors.New(response.ErrorExplanation)
	}
}

func (obj *Object) Wait(UUID uuid.UUID) error {
	return acks.ACKS.Wait(UUID)
}

func (obj *Object) AddLocal(format contracts.Format, data []byte) error {
	URL := fmt.Sprintf("https://%s/api/v1/database/create/%s", obj.client.API, format.ToString())
	response := network.Send(obj.client.Http, URL, "POST", data)

	logger.Log.Debug("object add", zap.String("URL", URL), zap.String("data", string(data)))

	if response.Success {
		return acks.ACKS.Ack(format.GetUUID())
	} else {
		return errors.New(response.ErrorExplanation)
	}
}

func (obj *Object) RemoveLocal(format contracts.Format) (bool, error) {
	prefix := format.ToString()

	if !format.Full() {
		// Append dot to the end of the format so that we delimit what we deleting from the kv-store
		prefix += "."
	}

	URL := fmt.Sprintf("https://%s/api/v1/database/keys/%s", obj.client.API, prefix)
	response := network.Send(obj.client.Http, URL, "DELETE", nil)

	logger.Log.Debug("object remove", zap.String("URL", URL))

	if response.Success {
		return true, nil
	} else {
		return false, errors.New(response.ErrorExplanation)
	}
}

func (obj *Object) Find(format contracts.Format) error {
	URL := fmt.Sprintf("https://%s/api/v1/database/get/%s", obj.client.API, format.ToString())
	response := network.Send(obj.client.Http, URL, "GET", nil)

	logger.Log.Debug("object find", zap.String("URL", URL))

	if response.Success {
		obj.Byte, _ = response.Data.MarshalJSON()
		obj.String = string(obj.Byte)

		err := json.Unmarshal(obj.Byte, &obj.Definition)

		if err != nil {
			logger.Log.Debug("failed to unmarshal json from object find to map[string]interface{}", zap.String("data", obj.String))
		}

		obj.changed = false
		obj.exists = true
	} else {
		if response.HttpStatus != http.StatusNotFound {
			return errors.New(response.ErrorExplanation)
		}
	}

	return nil
}

func (obj *Object) FindMany(format contracts.Format) (map[string]contracts.ObjectInterface, error) {
	var objects = make(map[string]contracts.ObjectInterface)

	URL := fmt.Sprintf("https://%s/api/v1/database/keys/%s", obj.client.API, format.ToString())
	response := network.Send(obj.client.Http, URL, "GET", nil)

	logger.Log.Debug("object find many", zap.String("URL", URL))

	if response.Success {
		if response.Data != nil {
			var keys []string

			bytes, _ := response.Data.MarshalJSON()
			err := json.Unmarshal(bytes, &keys)

			if err != nil {
				return nil, err
			}

			if format.GetType() == f.TYPE_FORMATED {
				for _, key := range keys {
					objTmp := New(obj.client, obj.User)
					err = objTmp.Find(f.NewFromString(key))

					if err != nil {
						return objects, err
					}

					objects[key] = objTmp
				}
			} else {
				for _, key := range keys {
					objTmp := New(obj.client, obj.User)
					err = objTmp.Find(f.NewUnformated(key, format.GetCategory()))

					if err != nil {
						return objects, err
					}

					objects[key] = objTmp
				}
			}
		}
	} else {
		return nil, errors.New(response.ErrorExplanation)
	}

	return objects, nil
}

func (obj *Object) Diff(definition []byte) bool {
	data := make(map[string]any)
	err := json.Unmarshal(definition, &data)

	if err != nil {
		return true
	}

	obj.Changelog, _ = diff.Diff(obj.Definition, data)

	if len(obj.Changelog) > 0 {
		obj.changed = true
	} else {
		obj.changed = false
	}

	return obj.changed
}

func (obj *Object) GetDiff() []diff.Change {
	return obj.Changelog
}

func (obj *Object) Exists() bool {
	return obj.exists
}

func (obj *Object) ChangeDetected() bool {
	return obj.changed
}

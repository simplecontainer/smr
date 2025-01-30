package secrets

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/r3labs/diff/v3"
	"github.com/simplecontainer/smr/pkg/acks"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

//go:generate mockgen -source=Interface.go -destination=mock/Interface.go

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

func (obj *Object) Propose(format contracts.Format, data []byte) error {
	URL := fmt.Sprintf("https://%s/api/v1/secrets/propose/%s/%s", obj.client.API, format.GetCategory(), format.ToString())
	response := SendRequest(obj.client.Http, URL, "POST", []byte(data))

	logger.Log.Debug("object add", zap.String("URL", URL), zap.String("data", string(data)))

	if response.Success {
		return nil
	} else {
		return errors.New(response.ErrorExplanation)
	}
}

func (obj *Object) Wait(format contracts.Format, data []byte) error {
	var wg sync.WaitGroup
	var errWait error

	go func() {
		wg.Add(1)
		errWait = acks.ACKS.Wait(format.GetUUID())
		wg.Done()
	}()

	err := obj.Propose(format, data)

	if err != nil {
		return err
	}

	wg.Wait()
	return errWait
}

func (obj *Object) AddLocal(format contracts.Format, data []byte) error {
	URL := fmt.Sprintf("https://%s/api/v1/secrets/create/%s", obj.client.API, format.ToString())
	response := SendRequest(obj.client.Http, URL, "POST", []byte(data))

	logger.Log.Debug("object add", zap.String("URL", URL), zap.String("data", string(data)))

	if response.Success {
		return nil
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

	URL := fmt.Sprintf("https://%s/api/v1/secrets/keys/%s", obj.client.API, prefix)
	response := SendRequest(obj.client.Http, URL, "DELETE", nil)

	logger.Log.Debug("object remove", zap.String("URL", URL))

	if response.Success {
		return true, nil
	} else {
		return false, errors.New(response.ErrorExplanation)
	}
}

func (obj *Object) Find(format contracts.Format) error {
	URL := fmt.Sprintf("https://%s/api/v1/secrets/get/%s", obj.client.API, format.ToString())
	response := SendRequest(obj.client.Http, URL, "GET", nil)

	logger.Log.Debug("object find", zap.String("URL", URL))

	if response.Success {
		b64decoded, _ := base64.StdEncoding.DecodeString(strings.Trim(string(response.Data), "\""))

		obj.Byte = b64decoded
		obj.String = string(b64decoded)
	} else {
		return errors.New(response.ErrorExplanation)
	}

	obj.changed = false
	obj.exists = true

	return nil
}

func (obj *Object) FindMany(format contracts.Format) (map[string]contracts.ObjectInterface, error) {
	var objects = make(map[string]contracts.ObjectInterface)

	URL := fmt.Sprintf("https://%s/api/v1/secrets/keys/%s", obj.client.API, format.ToString())
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

func (obj *Object) Diff(definition []byte) bool {
	data := make(map[string]any)
	err := json.Unmarshal([]byte(definition), &data)

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

package objects

import (
	json2 "encoding/json"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/simplecontainer/smr/pkg/acks"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/wI2L/jsondiff"
	"go.uber.org/zap"

	"net/http"

	"sync"
	"time"
)

//go:generate mockgen -source=Interface.go -destination=mock/Interface.go

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func New(client *client.Client, user *authentication.User) contracts.ObjectInterface {
	return &Object{
		Changelog: jsondiff.Patch{},
		client:    client,
		Byte:      make([]byte, 0),
		exists:    false,
		changed:   false,
		Created:   time.Now(),
		Updated:   time.Now(),
		User:      user,
	}
}

func (obj *Object) GetDefinition() map[string]any {
	var definition map[string]any
	json.Unmarshal(obj.Byte, &definition)

	return definition
}

func (obj *Object) GetDefinitionByte() []byte {
	return obj.Byte
}

func (obj *Object) Propose(format contracts.Format, data []byte) error {
	var URL string

	if format.GetType() == f.TYPE_FORMATED {
		URL = fmt.Sprintf("https://%s/api/v1/kind/propose/%s", obj.client.API, format.ToStringWithUUID())
	} else {
		URL = fmt.Sprintf("https://%s/api/v1/key/propose/%s", obj.client.API, format.ToStringWithUUID())
	}

	response := network.Send(obj.client.Http, URL, "POST", data)

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
	URL := fmt.Sprintf("https://%s/api/v1/kind/%s", obj.client.API, format.ToString())
	response := network.Send(obj.client.Http, URL, "POST", data)

	logger.Log.Debug("object add", zap.String("URL", URL), zap.String("data", string(data)))

	if response.Success {
		return nil
	} else {
		return errors.New(fmt.Sprintf("%s: %s", response.ErrorExplanation, string(response.Data)))
	}
}

func (obj *Object) RemoveLocal(format contracts.Format) (bool, error) {
	URL := fmt.Sprintf("https://%s/api/v1/kind/%s", obj.client.API, format.ToString())
	response := network.Send(obj.client.Http, URL, "DELETE", nil)

	logger.Log.Debug("object remove", zap.String("URL", URL))

	if response.Success {
		return true, nil
	} else {
		return false, errors.New(fmt.Sprintf("%s: %s", response.ErrorExplanation, string(response.Data)))
	}
}

func (obj *Object) AddLocalKey(key string, data []byte) error {
	URL := fmt.Sprintf("https://%s/api/v1/key/set/%s", obj.client.API, key)
	response := network.Send(obj.client.Http, URL, "POST", data)

	logger.Log.Debug("object add", zap.String("URL", URL), zap.String("data", string(data)))

	if response.Success {
		return nil
	} else {
		return errors.New(fmt.Sprintf("%s: %s", response.ErrorExplanation, string(response.Data)))
	}
}

func (obj *Object) RemoveLocalKey(key string) (bool, error) {
	URL := fmt.Sprintf("https://%s/api/v1/key/remove/%s", obj.client.API, key)
	response := network.Send(obj.client.Http, URL, "DELETE", nil)

	logger.Log.Debug("object remove", zap.String("URL", URL))

	if response.Success {
		return true, nil
	} else {
		return false, errors.New(fmt.Sprintf("%s: %s", response.ErrorExplanation, string(response.Data)))
	}
}

func (obj *Object) Find(format contracts.Format) error {
	URL := fmt.Sprintf("https://%s/api/v1/kind/%s", obj.client.API, format.ToString())
	response := network.Send(obj.client.Http, URL, "GET", nil)

	logger.Log.Debug("object find", zap.String("URL", URL))

	if response.Success {
		obj.Byte, _ = response.Data.MarshalJSON()

		obj.changed = false
		obj.exists = true
	} else {
		obj.exists = false

		if response.HttpStatus != http.StatusNotFound {
			return errors.New(fmt.Sprintf("%s: %s", response.ErrorExplanation, string(response.Data)))
		}
	}

	return nil
}

func (obj *Object) FindMany(format contracts.Format) ([]contracts.ObjectInterface, error) {
	var objects = make([]contracts.ObjectInterface, 0)

	URL := fmt.Sprintf("https://%s/api/v1/kind/%s", obj.client.API, format.ToString())
	response := network.Send(obj.client.Http, URL, "GET", nil)

	logger.Log.Debug("object find many", zap.String("URL", URL))

	if response.Success {
		if response.Data != nil {
			var tmp []json2.RawMessage
			err := json.Unmarshal(response.Data, &tmp)

			for _, j := range tmp {
				objTmp := New(obj.client, obj.User).(*Object)
				objTmp.Byte, _ = j.MarshalJSON()

				objects = append(objects, objTmp)
			}

			if err != nil {
				return nil, err
			}

			return objects, nil
		}
	} else {
		return nil, errors.New(fmt.Sprintf("%s: %s", response.ErrorExplanation, string(response.Data)))
	}

	return objects, nil
}

func (obj *Object) Diff(definition []byte) bool {
	objByte := obj.GetDefinitionByte()

	if len(objByte) == 0 {
		objByte = []byte(`{}`)
	}

	obj.Changelog, _ = jsondiff.CompareJSON(objByte, definition)

	if len(obj.Changelog) > 0 {
		obj.changed = true
	} else {
		obj.changed = false
	}

	return obj.changed
}

func (obj *Object) GetDiff() jsondiff.Patch {
	return obj.Changelog
}

func (obj *Object) Exists() bool {
	return obj.exists
}

func (obj *Object) ChangeDetected() bool {
	return obj.changed
}

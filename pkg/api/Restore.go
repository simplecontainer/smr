package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/manager"
	"github.com/simplecontainer/smr/pkg/objects"
	"io"
	"net/http"
)

func (api *Api) Restore(c *gin.Context) {
	var format *f.Format
	client, err := manager.GenerateHttpClient(api.Keys)

	if err != nil {
		c.JSON(http.StatusOK, httpcontract.ResponseOperator{
			Explanation:      "failed to generate client for the internal secure communication",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})
	}

	response := make(map[string]any, 0)

	format = f.New("containers")
	obj := objects.New(client)

	var objs map[string]*objects.Object
	objs, err = obj.FindMany(format)

	for name, object := range objs {
		b64decoded := make([]byte, 0)
		b64decoded, err = base64.StdEncoding.DecodeString(object.GetDefinitionString())

		sendRequest(client, "https://localhost:1443/api/v1/apply", string(b64decoded))
		response[name] = object
	}

	c.JSON(http.StatusOK, httpcontract.ResponseOperator{
		Explanation:      "here is the item list from the db",
		ErrorExplanation: "",
		Error:            true,
		Success:          false,
		Data:             response,
	})
}

func sendRequest(client *http.Client, URL string, data string) *httpcontract.ResponseImplementation {
	var req *http.Request
	var err error

	if len(data) > 0 {
		if err != nil {
			return &httpcontract.ResponseImplementation{
				HttpStatus:       0,
				Explanation:      "failed to marshal data for sending request",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
			}
		}

		req, err = http.NewRequest("POST", URL, bytes.NewBuffer([]byte(data)))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest("GET", URL, nil)
		req.Header.Set("Content-Type", "application/json")
	}

	if err != nil {
		return &httpcontract.ResponseImplementation{
			HttpStatus:       0,
			Explanation:      "failed to craft request",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}

	resp, err := client.Do(req)

	if err != nil {
		return &httpcontract.ResponseImplementation{
			HttpStatus:       0,
			Explanation:      "failed to connect to the smr-agent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &httpcontract.ResponseImplementation{
			HttpStatus:       0,
			Explanation:      "invalid response from the smr-agent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}

	var response httpcontract.ResponseImplementation
	err = json.Unmarshal(body, &response)

	if err != nil {
		return &httpcontract.ResponseImplementation{
			HttpStatus:       0,
			Explanation:      "failed to unmarshal body response from smr-agent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}

	return &response
}

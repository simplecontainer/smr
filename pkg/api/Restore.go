package api

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"github.com/simplecontainer/smr/pkg/objects"
	"io"
	"net/http"
)

func (api *Api) Restore(c *gin.Context) {
	var formatContainers *f.Format
	var formatGitops *f.Format

	data := make(map[string]any, 0)

	user := authentication.NewUser(c.Request.TLS)

	formatContainers = f.New("containers")
	formatGitops = f.New("gitops")
	obj := objects.New(api.Manager.Http.Get(user.Username), user)

	objsTmp, errTmp := obj.FindMany(formatContainers)

	if errTmp != nil {
		c.JSON(http.StatusInternalServerError, httpcontract.ResponseOperator{
			Explanation:      "",
			ErrorExplanation: errTmp.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	for name, object := range objsTmp {
		response := sendRequest(api.Manager.Http, user, "https://localhost:1443/api/v1/apply/containers", string(object.GetDefinitionByte()))

		if !response.Error {
			data[name] = string(object.GetDefinitionByte())
		} else {
			data[name] = response.ErrorExplanation
		}
	}

	objsTmp, errTmp = obj.FindMany(formatGitops)

	if errTmp != nil {
		c.JSON(http.StatusInternalServerError, httpcontract.ResponseOperator{
			Explanation:      "",
			ErrorExplanation: errTmp.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	for name, object := range objsTmp {
		response := sendRequest(api.Manager.Http, user, "https://localhost:1443/api/v1/apply/gitops", string(object.GetDefinitionByte()))

		if !response.Error {
			data[name] = string(object.GetDefinitionByte())
		} else {
			data[name] = response.ErrorExplanation
		}
	}

	c.JSON(http.StatusOK, httpcontract.ResponseOperator{
		Explanation:      "here is the item list from the db",
		ErrorExplanation: "",
		Error:            true,
		Success:          false,
		Data:             data,
	})
}

func sendRequest(client *client.Http, user *authentication.User, URL string, data string) *httpcontract.ResponseImplementation {
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

	resp, err := client.Get(user.Username).Http.Do(req)

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

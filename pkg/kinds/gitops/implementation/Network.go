package implementation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts"
	"io"
	"net/http"
)

func (gitops *Gitops) sendRequest(client *client.Http, user *authentication.User, URL string, data []byte) *contracts.Response {
	var req *http.Request
	var err error

	if len(data) > 0 {
		if err != nil {
			return &contracts.Response{
				HttpStatus:       0,
				Explanation:      "failed to marshal data for sending request",
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
			}
		}

		req, err = http.NewRequest("POST", URL, bytes.NewBuffer(data))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Owner", fmt.Sprintf("gitops.%s.%s", gitops.Definition.Meta.Group, gitops.Definition.Meta.Name))
	} else {
		req, err = http.NewRequest("GET", URL, nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Owner", fmt.Sprintf("gitops.%s.%s", gitops.Definition.Meta.Group, gitops.Definition.Meta.Name))
	}

	if err != nil {
		return &contracts.Response{
			HttpStatus:       0,
			Explanation:      "failed to craft request",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}

	resp, err := client.Get(user.Username).Http.Do(req)

	if err != nil {
		return &contracts.Response{
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
		return &contracts.Response{
			HttpStatus:       0,
			Explanation:      "invalid response from the smr-agent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}

	var response contracts.Response
	err = json.Unmarshal(body, &response)

	if err != nil {
		return &contracts.Response{
			HttpStatus:       0,
			Explanation:      "failed to unmarshal body response from smr-agent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		}
	}

	return &response
}

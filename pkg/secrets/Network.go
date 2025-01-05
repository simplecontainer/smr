package secrets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts"
	"io"
	"net/http"
)

func SendRequest(client *http.Client, URL string, method string, data []byte) *contracts.Response {
	var req *http.Request
	var marshaled []byte
	var err error

	if data != nil {
		req, err = http.NewRequest(method, URL, bytes.NewBuffer(data))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, URL, nil)
		req.Header.Set("Content-Type", "application/json")
	}

	if err != nil {
		return &contracts.Response{
			HttpStatus:       0,
			Explanation:      "failed to craft request",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	resp, err := client.Do(req)

	if err != nil {
		return &contracts.Response{
			HttpStatus:       0,
			Explanation:      "failed to connect to the smr-agent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
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
			Data:             nil,
		}
	}

	var response contracts.Response
	err = json.Unmarshal(body, &response)

	if err != nil {
		return &contracts.Response{
			HttpStatus:       resp.StatusCode,
			Explanation:      "failed to unmarshal body response from smr-agent",
			ErrorExplanation: generateResponse(URL, method, marshaled, body, err),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	response.HttpStatus = resp.StatusCode
	return &response
}

func generateResponse(URL string, method string, data []byte, body []byte, err error) string {
	debug := fmt.Sprintf("URL: %s METHOD: %s SEND_DATA: %s RESPONSE: %s", URL, method, string(data), string(body))
	return fmt.Sprintf("database returned malformed response - " + debug + "\n" + err.Error())
}

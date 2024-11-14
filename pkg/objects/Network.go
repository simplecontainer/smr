package objects

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts"
	"io"
	"net/http"
)

func SendRequest(client *http.Client, URL string, method string, data []byte) *contracts.ResponseOperator {
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
		return &contracts.ResponseOperator{
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
		return &contracts.ResponseOperator{
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
		return &contracts.ResponseOperator{
			HttpStatus:       0,
			Explanation:      "invalid response from the smr-agent",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	var response contracts.ResponseOperator
	err = json.Unmarshal(body, &response)

	if err != nil {
		return &contracts.ResponseOperator{
			HttpStatus:       0,
			Explanation:      "failed to unmarshal body response from smr-agent",
			ErrorExplanation: generateResponse(URL, method, marshaled, body),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	return &response
}

func generateResponse(URL string, method string, data []byte, body []byte) string {
	debug := fmt.Sprintf("URL: %s METHOD: %s SEND_DATA: %s RESPONSE: %s", URL, method, string(data), string(body))
	return fmt.Sprintf("failed to fetch object from the kv store: %s", debug)
}

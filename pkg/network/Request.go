package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/simplecontainer/smr/pkg/contracts"
	"io"
	"net/http"
)

func Send(client *http.Client, URL string, method string, data []byte) *contracts.Response {
	var req *http.Request
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
			Explanation:      "failed to connect to the node",
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
			Explanation:      "invalid response from the node",
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
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            false,
			Success:          false,
			Data:             body,
		}
	}

	response.HttpStatus = resp.StatusCode
	return &response
}

func Raw(client *http.Client, URL string, method string, data interface{}) (*http.Response, error) {
	var req *http.Request
	var err error

	if data != nil {
		var marshaled []byte
		marshaled, err = json.Marshal(data)

		switch v := data.(type) {
		case string:
			marshaled = []byte(v)
			break
		default:
			marshaled, err = json.Marshal(v)
		}

		if err != nil {
			return nil, err
		}

		req, err = http.NewRequest(method, URL, bytes.NewBuffer(marshaled))

		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, URL, nil)

		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func generateResponse(URL string, status int, method string, data []byte, body []byte, err error) string {
	debug := fmt.Sprintf("URL: %s RESPONSE_CODE: %d, METHOD: %s SEND_DATA: %s RESPONSE: %s", URL, status, method, string(data), string(body))
	return fmt.Sprintf("database returned malformed response - " + debug + "\n" + err.Error())
}

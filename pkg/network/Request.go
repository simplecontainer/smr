package network

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"io"
	"net/http"
	"time"
)

func Send(client *http.Client, URL string, method string, data []byte) *iresponse.Response {
	var req *http.Request
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if data != nil {
		req, err = http.NewRequestWithContext(ctx, method, URL, bytes.NewBuffer(data))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, URL, nil)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "smr")

	if err != nil {
		return &iresponse.Response{
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
		return &iresponse.Response{
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
		return &iresponse.Response{
			HttpStatus:       0,
			Explanation:      "invalid response from the node",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		}
	}

	var response iresponse.Response
	err = json.Unmarshal(body, &response)

	if err != nil {
		return &iresponse.Response{
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

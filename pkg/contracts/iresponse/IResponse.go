package iresponse

import "encoding/json"

type Response struct {
	HttpStatus       int
	Explanation      string
	ErrorExplanation string
	Error            bool
	Success          bool
	Data             json.RawMessage
}

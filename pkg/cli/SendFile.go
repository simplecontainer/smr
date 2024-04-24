package cli

import (
	"fmt"
	"github.com/imroc/req/v3"
	"log"
)

type Result struct {
	Data string `json:"data"`
}

func SendFile(URL string, jsonData string) {
	client := req.C().DevMode()
	var result Result

	resp, err := client.R().
		SetBodyJsonString(jsonData).
		SetSuccessResult(&result).
		Post(URL)
	if err != nil {
		log.Fatal(err)
	}

	if !resp.IsSuccessState() {
		fmt.Println("bad response status:", resp.Status)
		return
	}
}

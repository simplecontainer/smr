package cli

import (
	"fmt"
	"github.com/imroc/req/v3"
	"log"
	"smr/pkg/container"
)

func SendPs(URL string) map[string]map[string]*container.Container {
	client := req.C().DevMode()
	var data map[string]map[string]*container.Container

	resp, err := client.R().
		SetSuccessResult(&data).
		Get(URL)
	if err != nil {
		log.Fatal(err)
	}

	if !resp.IsSuccessState() {
		fmt.Println("bad response status:", resp.Status)
		return nil
	}

	return data
}

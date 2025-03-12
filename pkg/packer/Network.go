package packer

import (
	b64 "encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func Download(URL *url.URL) ([]byte, error) {
	path := fmt.Sprintf("/tmp/%s", b64.StdEncoding.EncodeToString([]byte(URL.String())))

	out, err := os.Create(path)
	defer out.Close()

	if err != nil {
		return nil, err
	}

	resp, err := http.Get(URL.String())
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Println("definition not found at specific URL")
		os.Exit(1)
	}

	return ReadYAMLFile(path)
}

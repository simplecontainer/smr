package helpers

import (
	"net/url"
	"strings"
)

func EnforceHTTPS(raw string) (*url.URL, error) {
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "http://" + raw // temporarily add a scheme to make it parsable
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	parsed.Scheme = "https"
	return parsed, nil
}

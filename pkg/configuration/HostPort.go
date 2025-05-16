package configuration

import "net/url"

func NewHostPort(rawURL string) (*HostPort, error) {
	URL, err := url.Parse(rawURL)

	if err != nil {
		return nil, err
	}

	return &HostPort{
		Host: URL.Host,
		Port: URL.Port(),
	}, nil
}

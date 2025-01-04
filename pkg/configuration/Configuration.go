package configuration

import (
	"net/url"
)

func NewConfig() *Configuration {
	return &Configuration{
		Certificates: &Certificates{},
	}
}

func (configuration *Configuration) GetDomainOrIP() string {
	URL, err := url.Parse(configuration.KVStore.URL)

	if err != nil {
		panic(err)
	}

	return URL.Hostname()
}

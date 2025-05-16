package dns

import (
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
)

func NewARecord() *ARecord {
	return &ARecord{
		Addresses: []string{},
	}
}

func (AR *ARecord) Append(ip string) {
	if !AR.contains(ip) {
		AR.Addresses = append(AR.Addresses, ip)
	}
}

func (AR *ARecord) contains(ip string) bool {
	for _, existingIP := range AR.Addresses {
		if existingIP == ip {
			return true
		}
	}
	return false
}

func (AR *ARecord) Remove(ip string) {
	var newAddresses []string
	for _, existingIP := range AR.Addresses {
		if existingIP != ip {
			newAddresses = append(newAddresses, existingIP)
		}
	}
	AR.Addresses = newAddresses
}

func (AR *ARecord) Fetch(client *clients.Http, user *authentication.User, domain string) ([]string, error) {
	format := f.New(static.SMR_PREFIX, static.CATEGORY_DNS, "dns", "internal", domain)
	obj := objects.New(client.Clients[user.Username], user)

	err := obj.Find(format)
	if err != nil || !obj.Exists() {
		return nil, ErrNotFound
	}

	var records []string
	err = json.Unmarshal(obj.GetDefinitionByte(), &records)
	if err != nil {
		return nil, ErrNotFound
	}

	return records, nil
}

func (AR *ARecord) ToJSON() ([]byte, error) {
	return json.Marshal(AR.Addresses)
}

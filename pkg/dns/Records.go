package dns

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
)

func NewARecord() *ARecord {
	return &ARecord{
		[]string{},
	}
}

func (AR *ARecord) Append(ip string) {
	for _, i := range AR.Addresses {
		if i == ip {
			return
		}
	}

	AR.Addresses = append(AR.Addresses, ip)
}

func (AR *ARecord) Remove(ip string) {
	for i, ARip := range AR.Addresses {
		if ARip == ip {
			AR.Addresses = append(AR.Addresses[:i], AR.Addresses[i+1:]...)
		}
	}
}

func (AR *ARecord) Fetch(client *client.Http, user *authentication.User, domain string) ([]string, error) {
	format := f.NewUnformated(fmt.Sprintf("dns.%s", domain), static.CATEGORY_PLAIN_STRING)
	obj := objects.New(client.Clients[user.Username], user)

	obj.Find(format)

	if obj.Exists() {
		records := make([]string, 0)

		err := json.Unmarshal(obj.GetDefinitionByte(), &records)

		if err != nil {
			return []string{}, ErrNotFound
		}

		return records, nil
	} else {
		return []string{}, ErrNotFound
	}
}

func (AR *ARecord) ToJson() ([]byte, error) {
	return json.Marshal(AR.Addresses)
}

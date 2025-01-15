package dns

import (
	"fmt"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
)

func (r *Records) Propose(domain string, ip string, action uint8) error {
	format := f.NewUnformated(fmt.Sprintf("dns.%s.%s", domain, r.Agent), static.CATEGORY_DNS_STRING)
	obj := objects.New(r.Client.Clients[r.User.Username], r.User)

	bytes, err := json.Marshal(Distributed{
		Domain: domain,
		IP:     ip,
		Action: action,
	})

	if err != nil {
		return err
	}

	err = obj.Propose(format, bytes)

	if err != nil {
		return err
	}

	return obj.Wait(format)
}

func (r *Records) Local(domain string, bytes []byte) error {
	format := f.NewUnformated(fmt.Sprintf("dns.%s.%s", domain, r.Agent), static.CATEGORY_DNS_STRING)
	obj := objects.New(r.Client.Clients[r.User.Username], r.User)

	return obj.AddLocal(format, bytes)
}

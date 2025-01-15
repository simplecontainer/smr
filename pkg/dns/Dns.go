package dns

import (
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/miekg/dns"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/distributed"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
	"net"
	"strings"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func New(agent string, client *client.Http, user *authentication.User) *Records {
	r := &Records{
		ARecords: make(map[string]*ARecord),
		Agent:    agent,
		Client:   client,
		User:     user,
		Updates:  make(chan distributed.KV),
	}

	return r
}

func (r *Records) ListenUpdates() {
	for {
		select {
		case data := <-r.Updates:
			d := Distributed{}
			err := json.Unmarshal(data.Val, &d)

			if err != nil {
				logger.Log.Error(err.Error())
				return
			}

			switch d.Action {
			case ADD_RECORD:
				err = r.AddARecord(d.Domain, d.IP)

				if err != nil {
					logger.Log.Error(err.Error())
				}
				break
			case REMOVE_RECORD:
				err = r.RemoveARecord(d.Domain, d.IP)

				if err != nil {
					logger.Log.Error(err.Error())
				}
				break
			}
			break
		}
	}
}

func (r *Records) AddARecord(domain string, ip string) error {
	_, ARecordexists := r.ARecords[domain]

	if !ARecordexists {
		r.ARecords[domain] = NewARecord()
	}

	r.ARecords[domain].Append(ip)

	bytes, err := json.Marshal(r.ARecords[domain])

	if err != nil {
		return err
	}

	return r.Local(domain, bytes)
}
func (r *Records) RemoveARecord(domain string, ip string) error {
	if r.ARecords[domain] != nil {
		ips := r.Find(domain)

		if len(ips) > 0 {
			_, ARecordexists := r.ARecords[domain]

			if ARecordexists {
				err := r.ARecords[domain].Remove(ip)

				if err != nil {
					return err
				}
			}

			r.ARecords[domain].Append(ip)

			bytes, err := json.Marshal(r.ARecords[domain])

			if err != nil {
				return err
			}

			return r.Local(domain, bytes)
		} else {
			return errors.New(fmt.Sprintf("ip %s not found for specifed domain %s", ip, domain))
		}
	} else {
		return errors.New(fmt.Sprintf("ip %s not found for specifed domain %s", ip, domain))
	}
}

func (r *Records) Find(domain string) []string {
	_, exists := r.ARecords[domain]

	if exists {
		return r.ARecords[domain].IPs
	} else {
		format := f.NewUnformated(fmt.Sprintf("dns.%s.%s", strings.TrimSuffix(domain, "."), r.Agent), static.CATEGORY_PLAIN_STRING)
		obj := objects.New(r.Client.Clients[r.User.Username], r.User)

		obj.Find(format)

		if obj.Exists() {
			records := make(map[string][]string, 0)

			err := json.Unmarshal(obj.GetDefinitionByte(), &records)

			if err != nil {
				logger.Log.Error(err.Error())
				return []string{}
			}

			return records["IPs"]
		} else {
			return []string{}
		}
	}
}

func ParseQuery(cache *Records, m *dns.Msg) error {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			ips := cache.Find(q.Name)

			if len(ips) > 0 {
				var rr dns.RR
				var err error

				for _, ip := range ips {
					rr, err = dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))

					if err == nil {
						m.Answer = append(m.Answer, rr)
					}
				}
			} else {
				config := dns.ClientConfig{
					Servers:  []string{"8.8.8.8"},
					Port:     "53",
					Ndots:    0,
					Timeout:  1,
					Attempts: 5,
				}

				c := new(dns.Client)

				ms := new(dns.Msg)
				ms.SetQuestion(dns.Fqdn(q.Name), q.Qtype)
				ms.RecursionDesired = true

				r, _, err := c.Exchange(ms, net.JoinHostPort(config.Servers[0], config.Port))

				if err != nil {
					return err
				}

				if r == nil {
					return errors.New("response empty")
				}

				if r.Rcode != dns.RcodeSuccess {
					return errors.New("request failed")
				}

				for _, a := range r.Answer {
					m.Answer = append(m.Answer, a)
				}
			}
		}
	}

	return nil
}

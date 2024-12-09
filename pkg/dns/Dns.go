package dns

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/objects"
	"net"
)

func New(agent string) *Records {
	r := &Records{
		ARecords: make(map[string]*ARecord),
		Agent:    agent,
	}

	return r
}

func (r *Records) AddARecord(domain string, ip string) {
	_, ARecordexists := r.ARecords[domain]

	if !ARecordexists {
		r.ARecords[domain] = NewARecord()
	}

	r.ARecords[domain].Append(ip)

	format := f.NewFromString(fmt.Sprintf("dns.%s.%s", domain, r.Agent))
	obj := objects.New(r.Client.Clients[r.User.Username], r.User)

	bytes, err := json.Marshal(r.ARecords[domain].IPs)

	if err != nil {
		logger.Log.Error(err.Error())
		return
	}

	obj.Add(format, string(bytes))
}
func (r *Records) RemoveARecord(domain string, ip string) error {
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

		format := f.NewFromString(fmt.Sprintf("dns.%s.%s", domain, r.Agent))
		obj := objects.New(r.Client.Clients[r.User.Username], r.User)

		bytes, err := json.Marshal(r.ARecords[domain].IPs)

		if err != nil {
			logger.Log.Error(err.Error())
			return err
		}

		obj.Add(format, string(bytes))

		return nil
	} else {
		return errors.New(fmt.Sprintf("ip %s not found for specifed domain %s", ip, domain))
	}
}

func (r *Records) Find(domain string) []string {
	_, exists := r.ARecords[domain]

	if exists {
		return r.ARecords[domain].IPs
	} else {
		format := f.NewFromString(fmt.Sprintf("dns.%s", domain))
		obj := objects.New(r.Client.Clients[r.User.Username], r.User)

		objs, err := obj.FindMany(format)

		if err != nil {
			logger.Log.Error(err.Error())
		}

		if len(objs) > 0 {
			records := make([]string, 0)

			for _, v := range objs {
				record := make([]string, 0)
				err = json.Unmarshal(v.GetDefinitionByte(), &record)

				fmt.Println(v.GetDefinitionString())

				if err != nil {
					logger.Log.Error(err.Error())
					continue
				}

				records = append(records, record...)
			}

			return records
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

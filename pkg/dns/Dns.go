package dns

import (
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/miekg/dns"
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/network/nameservers"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/smaps"
	"github.com/simplecontainer/smr/pkg/static"
	"net"
	"strings"
)

var (
	ErrNotFound = errors.New("record not found")
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func New(agent string, client *client.Http, user *authentication.User) *Records {
	ns, err := nameservers.NewfromResolvConf(false)

	if err != nil {
		panic(err)
	}

	r := &Records{
		ARecords:    smaps.New(),
		Client:      client,
		User:        user,
		Nameservers: ns.ToString(),
		Records:     make(chan KV.KV),
	}

	return r
}

func (r *Records) AddARecord(domain string, ip string) ([]byte, error) {
	var record = &ARecord{}
	tmp, ok := r.ARecords.Map.Load(domain)

	if ok {
		record = tmp.(*ARecord)
		record.Append(ip)
	} else {
		AR := NewARecord()
		AR.Append(ip)

		r.ARecords.Map.Store(domain, AR)
	}

	return record.ToJson()
}

func (r *Records) RemoveARecord(domain string, ip string) ([]byte, error) {
	var record = &ARecord{}
	tmp, ok := r.ARecords.Map.Load(domain)

	if ok {
		record = tmp.(*ARecord)

		record.Remove(ip)
	}

	return record.ToJson()
}

func (r *Records) Save(bytes []byte, domain string) error {
	format, _ := f.New(static.SMR_PREFIX, static.CATEGORY_DNS, "dns", "internal", domain)
	obj := objects.New(r.Client.Clients[r.User.Username], r.User)

	return obj.AddLocal(format, bytes)
}

func (r *Records) Find(domain string) ([]string, error) {
	var record = &ARecord{}
	tmp, ok := r.ARecords.Map.Load(strings.TrimSuffix(domain, "."))

	if ok {
		record = tmp.(*ARecord)
		return record.Addresses, nil
	} else {
		return record.Fetch(r.Client, r.User, strings.TrimSuffix(domain, "."))
	}
}

func ParseQuery(records *Records, m *dns.Msg) (*dns.Msg, error) {
	if strings.HasSuffix(m.Question[0].Name, ".private.") {
		return LookupLocal(records, m)
	} else {
		return LookupRemote(records, m)
	}
}

func LookupLocal(records *Records, m *dns.Msg) (*dns.Msg, error) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			addresses, err := records.Find(q.Name)

			if err == nil {
				var rr dns.RR
				var err error

				for _, ip := range addresses {
					rr, err = dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))

					if err == nil {
						m.Answer = append(m.Answer, rr)
					}
				}

				return m, nil
			} else {
				return m, err
			}
		default:
			return m, ErrNotFound
		}
	}

	return m, errors.New("questions empty")
}

func LookupRemote(records *Records, m *dns.Msg) (*dns.Msg, error) {
	config := dns.ClientConfig{
		Servers:  records.Nameservers,
		Port:     "53",
		Ndots:    0,
		Timeout:  5,
		Attempts: 5,
	}

	c := new(dns.Client)
	m.RecursionDesired = true

	var r *dns.Msg
	var err error

	r, _, err = c.Exchange(m, net.JoinHostPort(config.Servers[0], config.Port))

	if err != nil {
		return m, err
	}

	if r.Rcode != dns.RcodeSuccess {
		return r, errors.New("request failed")
	}

	return r.SetReply(m), nil
}

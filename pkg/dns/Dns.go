package dns

import (
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/miekg/dns"
	"github.com/simplecontainer/smr/pkg/KV"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/network/nameservers"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/smaps"
	"github.com/simplecontainer/smr/pkg/static"
	"net"
	"strings"
	"sync"
)

var (
	ErrNotFound = errors.New("record not found")
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func New(agent string, client *clients.Http, user *authentication.User) *Records {
	ns, err := nameservers.NewfromResolvConf(false)

	if err != nil {
		panic(fmt.Sprintf("failed to load nameservers: %v", err))
	}

	return &Records{
		ARecords:    smaps.New(),
		Client:      client,
		User:        user,
		Nameservers: ns.ToString(),
		Lock:        &sync.RWMutex{},
		Records:     make(chan KV.KV),
	}
}

func (r *Records) getRecord(domain string) *ARecord {
	r.Lock.RLock()
	defer r.Lock.RUnlock()
	record, _ := r.ARecords.Map.Load(domain)
	if record != nil {
		return record.(*ARecord)
	}
	return nil
}

func (r *Records) AddARecord(domain, ip string) ([]byte, error) {
	record := r.getRecord(domain)
	if record == nil {
		record = NewARecord()
		r.ARecords.Map.Store(domain, record)
	}
	record.Append(ip)

	return record.ToJSON()
}

func (r *Records) RemoveARecord(domain, ip string) ([]byte, error) {
	record := r.getRecord(domain)
	if record == nil {
		return nil, nil
	}

	record.Remove(ip)

	if len(record.Addresses) == 0 {
		return nil, nil
	}
	return record.ToJSON()
}

func (r *Records) Save(bytes []byte, domain string) error {
	format := f.New(static.SMR_PREFIX, static.CATEGORY_DNS, "dns", "internal", domain)
	obj := objects.New(r.Client.Clients[r.User.Username], r.User)
	return obj.AddLocal(format, bytes)
}

func (r *Records) Remove(bytes []byte, domain string) (bool, error) {
	format := f.New(static.SMR_PREFIX, static.CATEGORY_DNS, "dns", "internal", domain)
	obj := objects.New(r.Client.Clients[r.User.Username], r.User)
	return obj.RemoveLocal(format)
}

func (r *Records) Find(domain string) ([]string, error) {
	trimmedDomain := strings.TrimSuffix(domain, ".")
	record := r.getRecord(trimmedDomain)

	if record == nil {
		return record.Fetch(r.Client, r.User, trimmedDomain)
	}
	return record.Addresses, nil
}

func ParseQuery(records *Records, m *dns.Msg) (*dns.Msg, error) {
	if strings.HasSuffix(m.Question[0].Name, ".private.") {
		return LookupLocal(records, m)
	}

	return LookupRemote(records, m)
}

func LookupLocal(records *Records, m *dns.Msg) (*dns.Msg, error) {
	for _, q := range m.Question {
		if q.Qtype == dns.TypeA {
			addresses, err := records.Find(q.Name)
			if err != nil {
				return m, err
			}

			for _, ip := range addresses {
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
				if err != nil {
					return m, fmt.Errorf("failed to create RR: %v", err)
				}

				m.Answer = append(m.Answer, rr)
			}

			return m, nil
		}

		if q.Qtype == dns.TypeA {
			m.Answer = append(m.Answer, nil)
		}
	}

	return m, nil
}

func LookupRemote(records *Records, m *dns.Msg) (*dns.Msg, error) {
	config := dns.ClientConfig{
		Servers:  records.Nameservers,
		Port:     "53",
		Ndots:    0,
		Timeout:  5,
		Attempts: 5,
	}

	client := new(dns.Client)
	m.RecursionDesired = true

	r, _, err := client.Exchange(m, net.JoinHostPort(config.Servers[0], config.Port))
	if err != nil {
		return m, fmt.Errorf("failed to perform DNS exchange: %v", err)
	}

	if r.Rcode != dns.RcodeSuccess {
		return r, errors.New("request failed")
	}

	return r.SetReply(m), nil
}

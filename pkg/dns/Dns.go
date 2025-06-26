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
	ns, search, err := nameservers.NewfromResolvConf(false)

	if err != nil {
		panic(fmt.Sprintf("failed to load nameservers: %v", err))
	}

	r := &Records{
		ARecords:    smaps.New(),
		Client:      client,
		User:        user,
		Nameservers: ns.ToString(),
		Search:      search,
		Searcher:    NewTrie(),
		Lock:        &sync.RWMutex{},
		Records:     make(chan KV.KV),
	}

	r.Searcher.Insert(".private")

	for _, suffix := range r.Search {
		parsed := strings.Replace(fmt.Sprintf(".private.%s", suffix), "..", ".", 1)
		r.Searcher.Insert(parsed)
	}

	return r
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
	fmt.Println("finding", trimmedDomain)

	record := r.getRecord(trimmedDomain)

	if record == nil {
		return record.Fetch(r.Client, r.User, trimmedDomain)
	}
	return record.Addresses, nil
}

func ParseQuery(records *Records, m *dns.Msg) (*dns.Msg, int, error) {
	for _, q := range m.Question {
		prefix, local := records.Searcher.EndsWithSuffix(q.Name)

		if local {
			m.Authoritative = true
			RR, code, err := LookupLocal(records, prefix, q)

			if err != nil {
				return m, code, err
			}

			m.Answer = append(m.Answer, RR...)
		} else {
			remote, code, err := LookupRemote(records, m)

			if err != nil {
				return m, code, err
			}

			m.Answer = append(m.Answer, remote.Answer...)
		}
	}

	if len(m.Answer) == 0 {
		return m, dns.RcodeNameError, errors.New("answer records empty")
	}

	return m, dns.RcodeSuccess, nil
}

func LookupLocal(records *Records, prefix string, q dns.Question) ([]dns.RR, int, error) {
	switch q.Qtype {
	case dns.TypeA:
		addresses, err := records.Find(fmt.Sprintf("%s.private", prefix))
		if err != nil {
			return nil, dns.RcodeNameError, err
		}

		var RRs []dns.RR

		for _, ip := range addresses {
			rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
			if err != nil {
				return nil, dns.RcodeServerFailure, fmt.Errorf("failed to create RR: %v", err)
			}

			RRs = append(RRs, rr)
		}

		return RRs, dns.RcodeSuccess, nil
	case dns.TypeAAAA:
		// Don't return yet — just leave AAAA unanswered.
		// Optionally set RcodeNameError if AAAA is required.
	default:
		return nil, dns.RcodeNotImplemented, errors.New("unsupported record queried")
	}

	return nil, dns.RcodeSuccess, nil
}

func LookupRemote(records *Records, m *dns.Msg) (*dns.Msg, int, error) {
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
		return m, dns.RcodeServerFailure, fmt.Errorf("failed to perform DNS exchange: %v", err)
	}

	if r.Rcode != dns.RcodeSuccess {
		return r, r.Rcode, fmt.Errorf("server responded with %d code", r.Rcode)
	}

	return r.SetReply(m), r.Rcode, nil
}

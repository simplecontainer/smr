package dns

import (
	"fmt"
	"github.com/miekg/dns"
	"go.uber.org/zap"
	"net"
	"smr/pkg/logger"
)

func New() *Records {
	return &Records{}
}

func (r *Records) AddARecord(domain string, ip string) {
	if len(r.ARecords) > 0 {
		_, ARecordexists := r.ARecords[domain]

		if !ARecordexists {
			tmp := r.ARecords
			tmp[domain] = ARecord{
				map[string][]string{
					domain: {ip},
				},
				map[string][]string{
					domain: {},
				},
			}

			r.ARecords = tmp
		}

		_, domainexists := r.ARecords[domain].Domain[domain]

		if domainexists {
			var contains bool
			for _, i := range r.ARecords[domain].Domain[domain] {
				if i == ip {
					contains = true
				}
			}

			if !contains {
				logger.Log.Info("appending dns A record", zap.String("domain", domain), zap.String("ip", ip))
				r.ARecords[domain].Domain[domain] = append(r.ARecords[domain].Domain[domain], ip)
			}
		} else {
			logger.Log.Info("adding dns A record", zap.String("domain", domain), zap.String("ip", ip))

			tmp := ARecord{
				map[string][]string{
					domain: {ip},
				},
				map[string][]string{
					domain: {},
				},
			}

			r.ARecords[domain] = tmp
		}
	} else {
		tmp := map[string]ARecord{
			domain: {
				map[string][]string{
					domain: {ip},
				},
				map[string][]string{
					domain: {},
				},
			},
		}

		r.ARecords = tmp
	}
}
func (r *Records) RemoveARecord(domain string, ip string) bool {
	ips := r.Find(domain)

	for i, v := range ips {
		if v == ip {
			logger.Log.Info(fmt.Sprintf("removing %s address from %s domain", domain, v))
			ips = append(ips[:i], ips[i+1:]...)
		}
	}

	r.ARecords[domain].Domain[domain] = ips

	return true
}

func (r *Records) RemoveARecordQueue(domain string, ip string) bool {
	ips := r.Find(domain)

	for i, v := range ips {
		if v == ip {
			logger.Log.Info(fmt.Sprintf("added %s address for %s domain to delete queue", v, domain))
			ips = append(ips[:i], ips[i+1:]...)
		}
	}

	r.ARecords[domain].DomainDelete[domain] = ips

	return true
}

func (r *Records) Find(domain string) []string {
	arecords, exists := r.ARecords[domain]

	if exists {
		return arecords.Domain[domain]
	} else {
		return []string{}
	}
}

func (r *Records) FindDeleteQueue(domain string) []string {
	arecords, exists := r.ARecords[domain]

	if exists {
		return arecords.DomainDelete[domain]
	} else {
		return []string{}
	}
}

func (r *Records) ResetDeleteQueue(domain string) {
	arecords, exists := r.ARecords[domain]

	if exists {
		logger.Log.Info(fmt.Sprintf("cleared delete queue from domain %s", domain))
		arecords.DomainDelete[domain] = []string{}
	}
}

func ParseQuery(cache *Records, m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			logger.Log.Info("Querying dns server", zap.String("fqdn", q.Name))
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
				if r == nil {
					logger.Log.Warn("dns resolution error", zap.String("error", err.Error()))
					return
				}

				if r.Rcode != dns.RcodeSuccess {
					logger.Log.Warn("dns invalid answer name after A query", zap.String("domain", q.Name))
					return
				}

				for _, a := range r.Answer {
					m.Answer = append(m.Answer, a)
				}
			}
		}
	}
}

package dns

import (
	"fmt"
	"github.com/miekg/dns"
	"go.uber.org/zap"
	"log"
	"net"
	"smr/pkg/logger"
)

func ParseQuery(cache map[string]string, m *dns.Msg) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			log.Printf("Query for %s\n", q.Name)
			ip := cache[q.Name]
			if ip != "" {
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
				if err == nil {
					m.Answer = append(m.Answer, rr)
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
					logger.Log.Warn("*** error: %s\n", zap.String("error", err.Error()))
				}

				if r.Rcode != dns.RcodeSuccess {
					logger.Log.Warn(" *** invalid answer name after A query", zap.String("domain", q.Name))
				}

				for _, a := range r.Answer {
					m.Answer = append(m.Answer, a)
				}
			}
		}
	}
}

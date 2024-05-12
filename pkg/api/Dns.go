package api

import (
	mdns "github.com/miekg/dns"
	"github.com/qdnqn/smr/pkg/dns"
)

func (api *Api) HandleDns(w mdns.ResponseWriter, r *mdns.Msg) {
	m := new(mdns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case mdns.OpcodeQuery:
		dns.ParseQuery(api.DnsCache, m)
	}

	w.WriteMsg(m)
}

package api

import (
	mdns "github.com/miekg/dns"
	"github.com/simplecontainer/smr/implementations/container/shared"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/plugins"
)

func (api *Api) HandleDns(w mdns.ResponseWriter, r *mdns.Msg) {
	m := new(mdns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case mdns.OpcodeQuery:
		pl := plugins.GetPlugin(api.Manager.Config.Root, "container.so")
		sharedObj := pl.GetShared().(*shared.Shared)

		dns.ParseQuery(sharedObj.DnsCache, m)
	}

	w.WriteMsg(m)
}

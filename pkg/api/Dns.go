package api

import (
	"errors"
	mdns "github.com/miekg/dns"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/logger"
)

func (api *Api) HandleDns(w mdns.ResponseWriter, r *mdns.Msg) {
	m := new(mdns.Msg)
	m.SetReply(r)
	m.Compress = false

	defer func(w mdns.ResponseWriter) {
		err := w.Close()

		if err != nil {
			logger.Log.Error(err.Error())
		}
	}(w)

	switch r.Opcode {
	case mdns.OpcodeQuery:
		response, err := dns.ParseQuery(api.DnsCache, m)

		if err != nil && !errors.Is(err, dns.ErrNotFound) {
			err = w.WriteMsg(response)

			if err != nil {
				logger.Log.Error(err.Error())
			}

			return
		}
	}

	err := w.WriteMsg(m)

	if err != nil {
		logger.Log.Error(err.Error())
	}
}

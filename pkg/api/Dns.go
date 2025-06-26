package api

import (
	mdns "github.com/miekg/dns"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/logger"
)

func (a *Api) HandleDns(w mdns.ResponseWriter, m *mdns.Msg) {
	go func() {
		m.Compress = false

		defer func(w mdns.ResponseWriter) {
			err := w.Close()

			if err != nil {
				logger.Log.Error(err.Error())
			}
		}(w)

		r, code, err := dns.ParseQuery(a.DnsCache, m)
		m.Rcode = code

		if err != nil {
			logger.Log.Error(err.Error())
		}

		err = w.WriteMsg(r)

		if err != nil {
			logger.Log.Error(err.Error())
		}
	}()
}

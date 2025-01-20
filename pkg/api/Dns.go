package api

import (
	mdns "github.com/miekg/dns"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/logger"
)

func (api *Api) HandleDns(w mdns.ResponseWriter, m *mdns.Msg) {
	go func() {
		m.Compress = false

		defer func(w mdns.ResponseWriter) {
			err := w.Close()

			if err != nil {
				logger.Log.Error(err.Error())
			}
		}(w)

		r, err := dns.ParseQuery(api.DnsCache, m)

		if err != nil {
			m.Rcode = mdns.RcodeNameError
		}

		err = w.WriteMsg(r)

		if err != nil {
			logger.Log.Error(err.Error())
		}
	}()
}

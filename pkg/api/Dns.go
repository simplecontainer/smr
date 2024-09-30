package api

import (
	"github.com/gin-gonic/gin"
	mdns "github.com/miekg/dns"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/httpcontract"
	"net/http"
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

func (api *Api) ListDns(c *gin.Context) {
	c.JSON(http.StatusOK, httpcontract.ResponseImplementation{
		HttpStatus:       http.StatusBadRequest,
		Explanation:      "definitions available on the server",
		ErrorExplanation: "",
		Error:            true,
		Success:          false,
		Data:             api.DnsCache,
	})
}

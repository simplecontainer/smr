package api

import (
	"github.com/gin-gonic/gin"
	mdns "github.com/miekg/dns"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/dns"
	"github.com/simplecontainer/smr/pkg/network"
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
	c.JSON(http.StatusOK, contracts.Response{
		HttpStatus:       http.StatusBadRequest,
		Explanation:      "definitions available on the server",
		ErrorExplanation: "",
		Error:            true,
		Success:          false,
		Data:             network.ToJson(api.DnsCache),
	})
}

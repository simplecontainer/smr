package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hpcloud/tail"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/objects"
	"io"
	"net/http"
)

func (api *Api) Debug(c *gin.Context) {
	kind := c.Param("kind")
	group := c.Param("group")
	identifier := c.Param("identifier")
	follow := c.Param("follow")

	w := c.Writer
	header := w.Header()
	header.Set("Transfer-Encoding", "chunked")
	header.Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	format := f.New(kind, group, identifier, "object")
	obj := objects.New(api.Manager.Http.Clients[api.User.Username], api.User)

	obj.Find(format)

	request, err := common.NewRequest(kind)

	if err != nil {
		w.Write([]byte(err.Error()))
		w.(http.Flusher).Flush()
		w.CloseNotify()
	}

	err = request.Definition.FromJson(obj.GetDefinitionByte())

	if err != nil {
		w.Write([]byte(err.Error()))
		w.(http.Flusher).Flush()
		w.CloseNotify()
	}

	if request.Definition.GetRuntime().GetNode() != api.Cluster.Node.NodeID {
		var nodeName string

		fmt.Println(api.Cluster.Cluster.Nodes)

		for _, node := range api.Cluster.Cluster.Nodes {
			if node.NodeID == request.Definition.GetRuntime().GetNode() {
				nodeName = node.NodeName
			}
		}

		fmt.Println(api.Manager.Http.Clients)
		fmt.Println(nodeName)

		resp, err := network.Raw(api.Manager.Http.Clients[nodeName].Http, fmt.Sprintf("%s/api/v1/debug/%s/%s/%s/%s", api.Manager.Http.Clients[nodeName].API, kind, group, identifier, follow), http.MethodGet, nil)

		var bytes int
		buff := make([]byte, 512)

		for {
			bytes, err = resp.Body.Read(buff)

			if bytes == 0 || err == io.EOF {
				err = resp.Body.Close()

				if err != nil {
					logger.Log.Error(err.Error())
				}

				w.CloseNotify()
				break
			}

			_, err = w.Write(buff[:bytes])

			if err != nil {
				logger.Log.Error(err.Error())
				w.CloseNotify()
				break
			}

			w.(http.Flusher).Flush()
		}
	}

	t, err := tail.TailFile(fmt.Sprintf("/tmp/%s.%s.%s.log", kind, group, identifier),
		tail.Config{
			Follow: true,
		},
	)

	if err != nil {
		w.Write([]byte(err.Error()))
		w.(http.Flusher).Flush()
		w.CloseNotify()
		return
	}

	for line := range t.Lines {
		select {
		case <-w.CloseNotify():
			return
		default:
			w.Write([]byte(fmt.Sprintf("%s\n", line.Text)))
			w.(http.Flusher).Flush()
			w.CloseNotify()
		}
	}
}

package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hpcloud/tail"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/kinds/container/shared"
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/objects"
	"github.com/simplecontainer/smr/pkg/static"
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

	if kind == static.KIND_CONTAINER {
		container := api.KindsRegistry[static.KIND_CONTAINER].GetShared().(*shared.Shared).Registry.Find(group, identifier)

		if container == nil {
			err := network.StreamByte([]byte("container is not found"), w)

			if err != nil {
				logger.Log.Error(err.Error())
			}
		} else {
			if container.IsGhost() {
				client, ok := api.Manager.Http.Clients[container.GetRuntime().NodeName]

				if !ok {
					err := network.StreamByte([]byte("container is not found"), w)

					if err != nil {
						logger.Log.Error(err.Error())
					}
				} else {
					resp, err := network.Raw(client.Http, fmt.Sprintf("https://%s/api/v1/debug//%s/%s/%s/%s", api.Manager.Http.Clients[container.GetRuntime().NodeName].API, kind, group, identifier, follow), http.MethodGet, nil)

					if err != nil {
						err = network.StreamByte([]byte(err.Error()), w)

						if err != nil {
							logger.Log.Error(err.Error())
						}
					}

					err = network.StreamHttp(resp.Body, w)

					if err != nil {
						err = network.StreamByte([]byte(err.Error()), w)

						if err != nil {
							logger.Log.Error(err.Error())
						}
					}

					network.StreamClose(w)
				}
			} else {
				var t *tail.Tail
				var err error

				t, err = tail.TailFile(fmt.Sprintf("/tmp/%s.%s.%s.log", kind, group, identifier),
					tail.Config{
						Follow: true,
					},
				)

				if err != nil {
					err = network.StreamByte([]byte(err.Error()), w)

					if err != nil {
						logger.Log.Error(err.Error())
					}
				}

				for line := range t.Lines {
					select {
					case <-w.CloseNotify():
						return
					default:
						err = network.StreamByte([]byte(fmt.Sprintf("%s\n", line.Text)), w)

						if err != nil {
							logger.Log.Error(err.Error())
						}
					}
				}
			}
		}
	} else {
		format := f.New(kind, group, identifier, "object")
		obj := objects.New(api.Manager.Http.Clients[api.User.Username], api.User)

		obj.Find(format)

		if !obj.Exists() {
			err := network.StreamByte([]byte("object is not found"), w)

			if err != nil {
				logger.Log.Error(err.Error())
			}
		} else {

			request, err := common.NewRequest(kind)

			if err != nil {
				err = network.StreamByte([]byte(err.Error()), w)

				if err != nil {
					logger.Log.Error(err.Error())
				}
			} else {
				err = request.Definition.FromJson(obj.GetDefinitionByte())

				if err != nil {
					err = network.StreamByte([]byte(err.Error()), w)

					if err != nil {
						logger.Log.Error(err.Error())
					}
				} else {
					if request.Definition.GetRuntime().GetNode() != api.Cluster.Node.NodeID {
						var nodeName string

						for _, node := range api.Cluster.Cluster.Nodes {
							if node.NodeID == request.Definition.GetRuntime().GetNode() {
								nodeName = node.NodeName
							}
						}

						client, ok := api.Manager.Http.Clients[nodeName]

						if !ok {
							err := network.StreamByte([]byte("object is not found x"), w)

							if err != nil {
								logger.Log.Error(err.Error())
							}
						} else {
							var resp *http.Response
							resp, err = network.Raw(client.Http, fmt.Sprintf("https://%s/api/v1/debug//%s/%s/%s/%s", api.Manager.Http.Clients[nodeName].API, kind, group, identifier, follow), http.MethodGet, nil)

							err = network.StreamHttp(resp.Body, w)

							if err != nil {
								err = network.StreamByte([]byte(err.Error()), w)

								if err != nil {
									logger.Log.Error(err.Error())
								}
							}
						}
					} else {
						var t *tail.Tail
						t, err = tail.TailFile(fmt.Sprintf("/tmp/%s.%s.%s.log", kind, group, identifier),
							tail.Config{
								Follow: true,
							},
						)

						if err != nil {
							err = network.StreamByte([]byte(err.Error()), w)

							if err != nil {
								logger.Log.Error(err.Error())
							}
						}

						for line := range t.Lines {
							select {
							case <-w.CloseNotify():
								return
							default:
								err = network.StreamByte([]byte(fmt.Sprintf("%s\n", line.Text)), w)

								if err != nil {
									logger.Log.Error(err.Error())
								}
							}
						}
					}
				}
			}
		}

		network.StreamClose(w)
	}
}

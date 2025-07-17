package api

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/definitions/commonv1"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/static"
	"io"
	"net/http"
)

func (a *Api) Propose(c *gin.Context) {
	jsonData, err := io.ReadAll(c.Request.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
	} else {
		data := make(map[string]interface{})
		err = json.Unmarshal(jsonData, &data)

		if err != nil {
			c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
		} else {
			kind, ok := data["kind"].(string)

			if !ok {
				c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", errors.New("invalid definition sent"), nil))
				return
			}

			_, ok = a.KindsRegistry[kind]

			if !ok {
				c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", errors.New("invalid definition sent"), nil))
			} else {
				var request *common.Request
				request, err = common.NewRequest(kind)

				if err != nil {
					c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "", err, nil))
				} else {
					if err = request.Definition.FromJson(jsonData); err != nil {
						c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "invalid definition sent", err, nil))
						return
					}

					var valid bool
					valid, err = request.Definition.Validate()

					if !valid {
						c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "invalid definition sent", err, nil))
						return
					}

					if request.Definition.GetRuntime() == nil {
						request.Definition.SetRuntime(&commonv1.Runtime{
							Owner:    &commonv1.Owner{},
							Node:     a.Cluster.Node.NodeID,
							NodeName: a.Cluster.Node.NodeName,
						})
					} else {
						request.Definition.Definition.GetRuntime().SetNode(a.Cluster.Node.NodeID)
						request.Definition.Definition.GetRuntime().SetNodeName(a.Cluster.Node.NodeName)
					}

					var format f.Format

					switch c.Param("action") {
					case "apply":
						break
					case "state":
						request.Definition.GetState().AddOpt("action", static.STATE_KIND)
						break
					case "remove":
						request.Definition.GetState().AddOpt("action", static.REMOVE_KIND)
						break
					}

					var bytes []byte
					bytes, err = request.Definition.ToJSON()

					if err != nil {
						c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, "invalid definition sent", err, nil))
						return
					}

					format = f.New(static.SMR_PREFIX, static.CATEGORY_KIND, kind, request.Definition.GetMeta().Group, request.Definition.GetMeta().Name)

					a.Cluster.KVStore.Propose(format.ToStringWithUUID(), bytes, a.Manager.Config.KVStore.Node.NodeID)

					c.JSON(http.StatusOK, common.Response(http.StatusOK, static.RESPONSE_SCHEDULED, nil, nil))
				}
			}
		}
	}
}

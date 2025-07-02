package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/simplecontainer/smr/pkg/f"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"github.com/simplecontainer/smr/pkg/metrics"
	"github.com/simplecontainer/smr/pkg/network"
	"github.com/simplecontainer/smr/pkg/static"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net/http"
	"strings"
)

// @Success		200	{object}	  contracts.Response
// @Failure		400	{object}	  contracts.Response
// @Failure		404	{object}	  contracts.Response
// @Failure		500	{object}	  contracts.Response
// @Router			/kind/{prefix}/{category}/{kind} [get]
func (a *Api) ListState(c *gin.Context) {
	prefix := c.Param("prefix")
	version := c.Param("version")
	category := c.Param("category")
	kind := c.Param("kind")

	format := f.New(prefix, version, category, kind)
	opts := f.DefaultToStringOpts()
	opts.AddTrailingSlash = true
	response, err := a.Etcd.Get(c.Request.Context(), format.ToStringWithOpts(opts), clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
	} else {
		kinds := make([]json.RawMessage, 0)

		for _, kv := range response.Kvs {
			tmp := f.NewFromString(strings.TrimPrefix(string(kv.Key), "/"))
			format = f.New(prefix, version, static.CATEGORY_KIND, kind, tmp.GetGroup(), tmp.GetName())

			definition, err := a.Etcd.Get(c.Request.Context(), format.ToStringWithOpts(opts))

			if err != nil {
				c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
				return
			} else {
				if len(definition.Kvs) > 0 {
					var state map[string]json.RawMessage
					if err = json.Unmarshal(kv.Value, &state); err != nil {
						c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
					}

					state["Definition"] = definition.Kvs[0].Value

					combined, err := json.Marshal(state)

					if err != nil {
						c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
						return
					} else {
						kinds = append(kinds, combined)
					}
				}
			}
		}

		c.JSON(http.StatusOK, common.Response(http.StatusOK, "", nil, network.ToJSON(kinds)))
	}
}

// @Success		200	{object}	  contracts.Response
// @Failure		400	{object}	  contracts.Response
// @Failure		404	{object}	  contracts.Response
// @Failure		500	{object}	  contracts.Response
// @Router		/kind/{prefix}/{category}/{kind}/{group} [get]
func (a *Api) ListStateGroup(c *gin.Context) {
	prefix := c.Param("prefix")
	version := c.Param("version")
	category := c.Param("category")
	kind := c.Param("kind")
	group := c.Param("group")

	format := f.New(prefix, version, category, kind, group)
	opts := f.DefaultToStringOpts()
	opts.AddTrailingSlash = true
	response, err := a.Etcd.Get(c.Request.Context(), format.ToStringWithOpts(opts), clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
	} else {
		kinds := make([]json.RawMessage, 0)

		for _, kv := range response.Kvs {
			tmp := f.NewFromString(strings.TrimPrefix(string(kv.Key), "/"))
			format = f.New(prefix, version, static.CATEGORY_KIND, kind, tmp.GetGroup(), tmp.GetName())

			definition, err := a.Etcd.Get(c.Request.Context(), format.ToStringWithOpts(opts))

			if err != nil {
				c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
			} else {
				var state map[string]json.RawMessage
				if err = json.Unmarshal(kv.Value, &state); err != nil {
					c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
				}

				state["Definition"] = definition.Kvs[0].Value

				combined, err := json.Marshal(state)

				if err != nil {
					c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
					return
				} else {
					kinds = append(kinds, combined)
				}
			}
		}

		c.JSON(http.StatusOK, common.Response(http.StatusOK, "", nil, network.ToJSON(kinds)))
	}
}

// GetState godoc
//
//	@Summary		Get specific kind
//	@Description	get specific kind from the store
//	@Tags			database
//	@Produce		json
//
// @Success		200	{object}	  contracts.Response
// @Failure		400	{object}	  contracts.Response
// @Failure		404	{object}	  contracts.Response
// @Failure		500	{object}	  contracts.Response
// @Router		/kind/{prefix}/state/{kind}/{group}/{name}/{field} [get]
func (a *Api) GetState(c *gin.Context) {
	metrics.DatabaseGet.Increment()
	prefix := c.Param("prefix")
	version := c.Param("version")
	category := c.Param("category")
	kind := c.Param("kind")
	group := c.Param("group")
	name := c.Param("name")
	field := c.Param("field")

	format := f.New(prefix, version, category, kind, group, name, field)
	opts := f.DefaultToStringOpts()
	opts.AddTrailingSlash = true
	response, err := a.Etcd.Get(c.Request.Context(), format.ToStringWithOpts(opts))

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
	} else {
		if len(response.Kvs) == 0 {
			c.JSON(http.StatusNotFound, common.Response(http.StatusNotFound, "", errors.New("resource not found"), nil))
		} else {
			//var bytes json.RawMessage
			//bytes, err = json.RawMessage(response.Kvs[0].Value).MarshalJSON()

			bytes := response.Kvs[0].Value

			// Attach definition to the state output
			format = f.New(prefix, version, static.CATEGORY_KIND, kind, group, name)

			response, err = a.Etcd.Get(c.Request.Context(), format.ToStringWithOpts(opts))

			if err != nil {
				c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
			} else {
				var state map[string]json.RawMessage
				if err = json.Unmarshal(bytes, &state); err != nil {
					c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
				}

				state["Definition"] = response.Kvs[0].Value

				var combined []byte
				combined, err = json.Marshal(state)

				if err != nil {
					c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
				} else {
					c.JSON(http.StatusOK, common.Response(http.StatusOK, "", nil, combined))
				}
			}
		}
	}
}

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
	"go.etcd.io/etcd/api/v3/mvccpb"
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
	group := c.Param("group")

	format := f.New(prefix, version, category, kind, group)
	opts := f.DefaultToStringOpts()
	opts.AddTrailingSlash = true
	response, err := a.Etcd.Get(c.Request.Context(), format.ToStringWithOpts(opts), clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))

	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	states, err := a.join(c, response.Kvs, prefix, version, kind)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "", nil, network.ToJSON(states)))
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
		return
	}

	if len(response.Kvs) == 0 {
		c.JSON(http.StatusNotFound, common.Response(http.StatusNotFound, "", errors.New("resource not found"), nil))
		return
	}

	state, err := a.join(c, response.Kvs, prefix, version, kind)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, "", err, nil))
		return
	}

	c.JSON(http.StatusOK, common.Response(http.StatusOK, "", nil, state[0]))
}

func (a *Api) join(c *gin.Context, kvs []*mvccpb.KeyValue, prefix, version, kind string) ([]json.RawMessage, error) {
	kinds := make([]json.RawMessage, 0)
	opts := f.DefaultToStringOpts()
	opts.AddTrailingSlash = true

	for _, kv := range kvs {
		tmp := f.NewFromString(strings.TrimPrefix(string(kv.Key), "/"))

		combined, err := a.append(c, kv.Value, prefix, version, kind, tmp.GetGroup(), tmp.GetName())
		if err != nil {
			kinds = append(kinds, kv.Value)
			continue
		}

		kinds = append(kinds, combined)
	}

	return kinds, nil
}

func (a *Api) append(c *gin.Context, stateValue []byte, prefix, version, kind, group, name string) (json.RawMessage, error) {
	format := f.New(prefix, version, static.CATEGORY_KIND, kind, group, name)
	opts := f.DefaultToStringOpts()
	opts.AddTrailingSlash = true

	definition, err := a.Etcd.Get(c.Request.Context(), format.ToStringWithOpts(opts))
	if err != nil {
		return nil, err
	}

	if len(definition.Kvs) == 0 {
		return nil, errors.New("definition not found")
	}

	var state map[string]json.RawMessage
	if err = json.Unmarshal(stateValue, &state); err != nil {
		return nil, err
	}

	state["Definition"] = definition.Kvs[0].Value

	combined, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}

	return combined, nil
}

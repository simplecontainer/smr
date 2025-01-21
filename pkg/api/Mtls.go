package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"net/http"
	"path/filepath"
	"strings"
)

func (api *Api) CreateUser(c *gin.Context) {
	user := authentication.NewUser(c.Request.TLS)
	path, err := user.CreateUser(api.Keys, api.Config.NodeName, c.Param("username"), c.Param("domain"), c.Param("externalIP"))

	if err == nil {
		var httpClient *http.Client
		httpClient, err = client.GenerateHttpClient(api.Keys.CA, api.Keys.Clients[c.Param("username")])

		if err != nil {
			c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, fmt.Sprintf("failed to create user credentials for: %s", filepath.Clean(c.Param("username"))), nil, nil))
			return
		}

		api.Manager.Http.Append(c.Param("username"), &client.Client{
			API:  fmt.Sprintf("%s:%s", c.Param("domain"), api.Config.HostPort.Port),
			Http: httpClient,
		})

		c.JSON(http.StatusOK, common.Response(http.StatusOK, fmt.Sprintf("user created, run: cat %s", strings.Replace(path, "/home/smr-agent", "$HOME", 1)), nil, nil))
	} else {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, fmt.Sprintf("failed to create user credentials for: %s", filepath.Clean(c.Param("username"))), nil, nil))

	}
}

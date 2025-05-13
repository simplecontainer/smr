package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/clients"
	"github.com/simplecontainer/smr/pkg/kinds/common"
	"net/http"
	"path/filepath"
	"strings"
)

func (a *Api) CreateUser(c *gin.Context) {
	user := authentication.NewUser(c.Request.TLS)
	path, err := user.CreateUser(a.Keys, a.Config.NodeName, c.Param("username"), c.Param("domain"), c.Param("externalIP"))

	if err == nil {
		var httpClient *http.Client
		httpClient, err = clients.GenerateHttpClient(a.Keys.CA, a.Keys.Clients[c.Param("username")])

		if err != nil {
			c.JSON(http.StatusInternalServerError, common.Response(http.StatusInternalServerError, fmt.Sprintf("failed to create user credentials for: %s", filepath.Clean(c.Param("username"))), nil, nil))
			return
		}

		a.Manager.Http.Append(c.Param("username"), &clients.Client{
			API:  fmt.Sprintf("%s:%s", c.Param("domain"), a.Config.HostPort.Port),
			Http: httpClient,
		})

		c.JSON(http.StatusOK, common.Response(http.StatusOK, fmt.Sprintf("user created, run: cat %s", strings.Replace(path, a.Config.Environment.Container.Home, "$HOME", 1)), nil, nil))
	} else {
		c.JSON(http.StatusBadRequest, common.Response(http.StatusBadRequest, fmt.Sprintf("failed to create user credentials for: %s", filepath.Clean(c.Param("username"))), nil, nil))

	}
}

package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/authentication"
	"github.com/simplecontainer/smr/pkg/client"
	"github.com/simplecontainer/smr/pkg/contracts"
	"net/http"
	"path/filepath"
	"strings"
)

func (api *Api) CreateUser(c *gin.Context) {
	user := authentication.NewUser(c.Request.TLS)
	path, err := user.CreateUser(api.Keys, api.Config.Node, c.Param("username"), c.Param("domain"), c.Param("externalIP"))

	if err == nil {
		var httpClient *http.Client
		httpClient, _, err = client.GenerateHttpClient(api.Keys.CA, api.Keys.Clients[c.Param("username")])

		if err != nil {
			c.JSON(http.StatusInternalServerError, contracts.Response{
				HttpStatus:       http.StatusBadRequest,
				Explanation:      fmt.Sprintf("failed to create user credentials for: %s", filepath.Clean(c.Param("username"))),
				ErrorExplanation: err.Error(),
				Error:            true,
				Success:          false,
			})

			return
		}

		api.Manager.Http.Append(c.Param("username"), &client.Client{
			API:  fmt.Sprintf("%s:%s", c.Param("domain"), api.Config.HostPort.Port),
			Http: httpClient,
		})

		c.JSON(http.StatusOK, contracts.Response{
			HttpStatus:       http.StatusOK,
			Explanation:      fmt.Sprintf("user created, run: cat %s", strings.Replace(path, "/home/smr-agent", "$HOME", 1)),
			ErrorExplanation: "",
			Error:            true,
			Success:          false,
		})
	} else {
		c.JSON(http.StatusBadRequest, contracts.Response{
			HttpStatus:       http.StatusBadRequest,
			Explanation:      fmt.Sprintf("failed to create user credentials for: %s", filepath.Clean(c.Param("username"))),
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
		})
	}
}

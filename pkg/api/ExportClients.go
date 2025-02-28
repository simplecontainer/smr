package api

import (
	"encoding/hex"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/contracts/iresponse"
	"github.com/simplecontainer/smr/pkg/encrypt"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/network"
	"net/http"
)

func (api *Api) ExportClients(c *gin.Context) {
	bytes, err := json.Marshal(api.Keys)

	if err != nil {
		c.JSON(http.StatusInternalServerError, &iresponse.Response{
			HttpStatus:       http.StatusInternalServerError,
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	var ciphertext string
	ciphertext, err = encrypt.Encrypt(string(bytes), hex.EncodeToString(api.Keys.Clients[api.User.Username].PrivateKeyBytes[:32]))

	encrypted := keys.Encrypted{Keys: ciphertext}

	if err != nil {
		c.JSON(http.StatusInternalServerError, &iresponse.Response{
			HttpStatus:       http.StatusInternalServerError,
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	c.JSON(http.StatusOK, &iresponse.Response{
		HttpStatus:       http.StatusOK,
		Explanation:      "Client certificates exported with success",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             network.ToJson(encrypted),
	})
}

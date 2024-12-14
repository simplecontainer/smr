package api

import (
	"encoding/hex"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/simplecontainer/smr/pkg/contracts"
	"github.com/simplecontainer/smr/pkg/encrypt"
	"github.com/simplecontainer/smr/pkg/keys"
	"github.com/simplecontainer/smr/pkg/network"
	"net/http"
)

func (api *Api) CA(c *gin.Context) {
	bytes, err := json.Marshal(api.Keys)

	if err != nil {
		c.JSON(http.StatusInternalServerError, &contracts.Response{
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
		c.JSON(http.StatusInternalServerError, &contracts.Response{
			HttpStatus:       http.StatusInternalServerError,
			Explanation:      "",
			ErrorExplanation: err.Error(),
			Error:            true,
			Success:          false,
			Data:             nil,
		})

		return
	}

	c.JSON(http.StatusOK, &contracts.Response{
		HttpStatus:       http.StatusOK,
		Explanation:      "CA exported with success",
		ErrorExplanation: "",
		Error:            false,
		Success:          true,
		Data:             network.ToJson(encrypted),
	})
}

package middleware

import (
	"github.com/gin-gonic/gin"
)

func TLSAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the tls is still valid
	}
}

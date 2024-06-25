package implementations

import (
	"github.com/gin-gonic/gin"
	"github.com/qdnqn/smr/pkg/httpcontract"
	"github.com/qdnqn/smr/pkg/manager"
)

// Plugin contracts
type Implementation interface {
	Apply(*manager.Manager, []byte, *gin.Context) (httpcontract.ResponseImplementation, error)
	Delete(*manager.Manager, []byte, *gin.Context) (httpcontract.ResponseImplementation, error)
}

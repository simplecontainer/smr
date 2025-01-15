package helpers

import (
	"github.com/simplecontainer/smr/pkg/logger"
	"github.com/simplecontainer/smr/pkg/static"
	"go.uber.org/zap"
)

func Category(category string) int {
	switch category {
	case "object":
		return static.CATEGORY_OBJECT
	case "etcd":
		return static.CATEGORY_ETCD
	case "secret":
		return static.CATEGORY_SECRET
	case "plain":
		return static.CATEGORY_PLAIN
	case "dns":
		return static.CATEGORY_DNS
	default:
		logger.Log.Error("invalid category sent", zap.String("category", category))
		return static.CATEGORY_INVALID
	}
}

func CategoryString(category int) string {
	switch category {
	case static.CATEGORY_OBJECT:
		return static.CATEGORY_OBJECT_STRING
	case static.CATEGORY_ETCD:
		return static.CATEGORY_ETCD_STRING
	case static.CATEGORY_SECRET:
		return static.CATEGORY_SECRET_STRING
	case static.CATEGORY_PLAIN:
		return static.CATEGORY_PLAIN_STRING
	case static.CATEGORY_DNS:
		return static.CATEGORY_DNS_STRING
	default:
		return static.CATEGORY_INVALID_STRING
	}
}

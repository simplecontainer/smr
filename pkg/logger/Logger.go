package logger

import (
	"go.uber.org/zap"
)

var Log *zap.Logger

func NewLogger() *zap.Logger {
	log, _ := zap.NewProduction()
	return log
}

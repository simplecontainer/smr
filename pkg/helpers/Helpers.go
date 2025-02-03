package helpers

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetLogLevel(level string) zapcore.Level {
	switch level {
	case "error":
		return zap.ErrorLevel
	case "warning":
		return zap.WarnLevel
	case "info":
		return zap.InfoLevel
	case "debug":
		return zap.DebugLevel
	}

	return zap.InfoLevel
}

func SplitClean(c rune) bool {
	return c == ','
}

package helpers

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetLogLevel(level string) zapcore.Level {
	switch level {
	case "error":
		return zap.ErrorLevel
		break
	case "warning":
		return zap.WarnLevel
		break
	case "info":
		return zap.InfoLevel
		break
	case "debug":
		return zap.DebugLevel
		break
	}

	return zap.InfoLevel
}

func SplitClean(c rune) bool {
	return c == ','
}

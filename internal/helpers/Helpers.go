package helpers

import (
	"fmt"
	"github.com/manifoldco/promptui"
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

func Confirm(message string) bool {
	ask := promptui.Select{
		Label: fmt.Sprintf("%s [y/n]", message),
		Items: []string{"y", "n"},
	}

	_, result, err := ask.Run()
	if err != nil {
		// if err provide simple yes no
		return false
	}

	return result == "y"
}

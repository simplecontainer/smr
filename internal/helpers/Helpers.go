package helpers

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
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

func CliMask(condition bool, textTrue string, textFalse string) string {
	if condition {
		return textTrue
	} else {
		return textFalse
	}
}

func CliRemoveComa(text string) string {
	str, _ := strings.CutSuffix(text, ", ")
	return str
}

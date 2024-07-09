package helpers

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
)

func Confirm(message string) bool {
	ask := promptui.Select{
		Label: fmt.Sprintf("%s [y/n]", message),
		Items: []string{"y", "n"},
	}

	_, result, err := ask.Run()
	if err != nil {
		log.Fatal("Prompt failed", err)
	}

	return result == "y"
}

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

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

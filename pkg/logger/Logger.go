package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"time"
)

var Log *zap.Logger

func NewLogger(logLevel string, outputStdout []string, outputStderr []string) *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	atomicLevel, err := zap.ParseAtomicLevel(logLevel)

	if err != nil {
		panic(err)
	}

	config := zap.Config{
		Level:             atomicLevel,
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths:       outputStdout,
		ErrorOutputPaths:  outputStderr,
		InitialFields:     map[string]interface{}{},
	}

	return zap.Must(config.Build())
}

func NewLoggerFile(dir string, name string, logLevel string) *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	atomicLevel, err := zap.ParseAtomicLevel(logLevel)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	log := filepath.Join(dir, fmt.Sprintf("%s.log", name))

	_, err = os.Create(log)
	if err != nil {
		panic(err)
	}

	config := zap.Config{
		Level:             atomicLevel,
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			log,
		},
		ErrorOutputPaths: []string{
			log,
		},
		InitialFields: map[string]interface{}{},
	}

	return zap.Must(config.Build())
}

func CreateOrRotate(path string) error {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()
		return nil
	} else if err != nil {
		return fmt.Errorf("error checking file: %w", err)
	}

	dir := filepath.Dir(path)
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	nameWithoutExt := filename[:len(filename)-len(ext)]

	timestamp := time.Now().Format("20060102_150405")
	newPath := filepath.Join(dir, fmt.Sprintf("%s_%s%s", nameWithoutExt, timestamp, ext))

	err = os.Rename(path, newPath)
	if err != nil {
		return fmt.Errorf("failed to rotate file: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create new file after rotation: %w", err)
	}
	defer file.Close()

	return nil
}

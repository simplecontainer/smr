package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

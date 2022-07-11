package qlog

import (
	"log"
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Zap() *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	// add colorization
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// we only want stack trace enabled for panic level and above
	logger, err := cfg.Build(zap.AddStacktrace(zapcore.PanicLevel))
	if err != nil {
		log.Fatalf("can't initialize logger: %v\n", err)
	}
	return logger
}

func Logger(t *testing.T) *zap.SugaredLogger {
	t.Helper()
	loggingEnabled := os.Getenv("LOG")
	if loggingEnabled == "" {
		return zap.NewNop().Sugar()
	}
	cfg := zap.NewDevelopmentConfig()
	// add colorization
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// we only want stack trace enabled for panic level and above
	logger, err := cfg.Build(zap.AddStacktrace(zapcore.PanicLevel))
	if err != nil {
		t.Fatalf("cannot initialize logger: %v", err)
	}
	return logger.Sugar()
}

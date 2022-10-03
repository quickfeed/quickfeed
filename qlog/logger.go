package qlog

import (
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Prod() *zap.SugaredLogger {
	cfg := zap.NewDevelopmentConfig()
	// cfg := zap.NewProductionConfig()
	// add colorization
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// we only want stack trace enabled for panic level and above
	logger, err := cfg.Build(zap.AddStacktrace(zapcore.PanicLevel))
	if err != nil {
		panic(err)
	}
	return logger.Sugar()
}

func Zap() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	// add colorization
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// we only want stack trace enabled for panic level and above
	return cfg.Build(zap.AddStacktrace(zapcore.PanicLevel))
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

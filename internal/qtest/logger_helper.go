package qtest

import (
	"os"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Logger(t *testing.T) *zap.SugaredLogger {
	t.Helper()
	if os.Getenv("LOG") == "" {
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

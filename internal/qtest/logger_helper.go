package qtest

import (
	"os"
	"testing"

	"go.uber.org/zap"
)

func Logger(t *testing.T) *zap.SugaredLogger {
	t.Helper()
	if os.Getenv("LOG") == "" {
		return zap.NewNop().Sugar()
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}
	return logger.Sugar()
}

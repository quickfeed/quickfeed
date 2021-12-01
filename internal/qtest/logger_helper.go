package qtest

import (
	"testing"

	"go.uber.org/zap"
)

func Logger(t *testing.T) *zap.SugaredLogger {
	t.Helper()
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal(err)
	}
	return logger.Sugar()
}

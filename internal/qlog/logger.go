package qlog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Zap() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	// add colorization
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// we only want stack trace enabled for panic level and above
	return cfg.Build(zap.AddStacktrace(zapcore.PanicLevel))
}

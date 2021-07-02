package log

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Zap(verbose bool) *zap.Logger {
	if verbose {
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
	return zap.NewNop()
}

package log

import (
	"log"

	"github.com/autograde/quickfeed/database"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Zap(verbose bool) *zap.Logger {
	if verbose {
		cfg := zap.NewDevelopmentConfig()
		// database logging is only enabled if the LOGDB environment variable is set
		cfg = database.GormLoggerConfig(cfg)
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

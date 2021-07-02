package database

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewGORMLogger returns a zap-based logger for GORM. A logger instance is only returned
// if the LOGDB environment variable is set to a specific level. This logger is not
// recommended for production due to the high volume of SQL queries being performed and
// the associated noise in the logs; it is mainly useful for debugging database issues.
// If LOGDB is not set, the discard logger is returned.
func NewGORMLogger(zapLogger *zap.Logger) gormlogger.Interface {
	var level gormlogger.LogLevel
	switch os.Getenv("LOGDB") {
	case "":
		return gormlogger.Discard
	case "1":
		level = gormlogger.Silent
	case "2":
		level = gormlogger.Error
	case "3":
		level = gormlogger.Warn
	case "4":
		level = gormlogger.Info
	}
	return Logger{
		ZapLogger:                 zapLogger,
		LogLevel:                  level,
		SlowThreshold:             100 * time.Millisecond,
		IgnoreRecordNotFoundError: false,
	}
}

// Logger is an adaption of gorm.Logger that uses the zap logger.
// The logger is based on code from moul.io/zapgorm2.
type Logger struct {
	ZapLogger                 *zap.Logger
	LogLevel                  gormlogger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

func (l Logger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return Logger{
		ZapLogger:                 l.ZapLogger,
		SlowThreshold:             l.SlowThreshold,
		LogLevel:                  level,
		IgnoreRecordNotFoundError: l.IgnoreRecordNotFoundError,
	}
}

func (l Logger) Info(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Info {
		return
	}
	l.logger().Sugar().Debugf(str, args...)
}

func (l Logger) Warn(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Warn {
		return
	}
	l.logger().Sugar().Warnf(str, args...)
}

func (l Logger) Error(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Error {
		return
	}
	l.logger().Sugar().Errorf(str, args...)
}

func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= 0 {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		l.logger().Error("trace", zap.Error(err), zap.Duration("elapsed", elapsed), zap.Int64("rows", rows), zap.String("sql", sql))
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		l.logger().Warn("trace", zap.Duration("elapsed", elapsed), zap.Int64("rows", rows), zap.String("sql", sql))
	case l.LogLevel >= gormlogger.Info:
		sql, rows := fc()
		l.logger().Debug("trace", zap.Duration("elapsed", elapsed), zap.Int64("rows", rows), zap.String("sql", sql))
	}
}

var (
	gormPackage      = filepath.Join("gorm.io", "gorm")
	quickfeedPackage = filepath.Join("github.com", "autograde", "quickfeed", "log")
)

func (l Logger) logger() *zap.Logger {
	for i := 2; i < 15; i++ {
		_, file, _, ok := runtime.Caller(i)
		switch {
		case !ok:
		case strings.HasSuffix(file, "_test.go"):
		case strings.Contains(file, gormPackage):
		case strings.Contains(file, quickfeedPackage):
		default:
			return l.ZapLogger.WithOptions(zap.AddCallerSkip(i))
		}
	}
	return l.ZapLogger
}

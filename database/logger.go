package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog/label"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// sqlLabel is the attribute holding the traced SQL statement.
const sqlLabel = "sql"

// NewGORMLogger returns a slog-based logger for GORM. A logger instance is only returned
// if the LOGDB environment variable is set to a specific level. This logger is not
// recommended for production due to the high volume of SQL queries being performed and
// the associated noise in the logs; it is mainly useful for debugging database issues.
// If LOGDB is not set, the discard logger is returned.
//
// The traced SQL statements include the bound column values, which may contain
// sensitive data, e.g., a user's refresh token. LOGDB must therefore only be
// enabled for local debugging, never in production. See doc/gorm-issues.md.
func NewGORMLogger(logger *slog.Logger) gormlogger.Interface {
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
		Logger:                    logger,
		LogLevel:                  level,
		SlowThreshold:             100 * time.Millisecond,
		IgnoreRecordNotFoundError: false,
	}
}

// Logger adapts slog to GORM's logger interface.
type Logger struct {
	Logger                    *slog.Logger
	LogLevel                  gormlogger.LogLevel
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

func (l Logger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return Logger{
		Logger:                    l.Logger,
		SlowThreshold:             l.SlowThreshold,
		LogLevel:                  level,
		IgnoreRecordNotFoundError: l.IgnoreRecordNotFoundError,
	}
}

func (l Logger) Info(_ context.Context, message string, args ...any) {
	if l.LogLevel < gormlogger.Info {
		return
	}
	l.Logger.Debug("database info", "message", expand(message, args...))
}

func (l Logger) Warn(_ context.Context, message string, args ...any) {
	if l.LogLevel < gormlogger.Warn {
		return
	}
	l.Logger.Warn("database warning", "message", expand(message, args...))
}

func (l Logger) Error(_ context.Context, message string, args ...any) {
	if l.LogLevel < gormlogger.Error {
		return
	}
	l.Logger.Error("database error", "message", expand(message, args...))
}

func (l Logger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= 0 {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		l.Logger.Error("database trace", sqlLabel, sql, "rows", rows, label.Duration, elapsed, label.Error, err)
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		l.Logger.Warn("slow database trace", sqlLabel, sql, "rows", rows, label.Duration, elapsed)
	case l.LogLevel >= gormlogger.Info:
		sql, rows := fc()
		l.Logger.Debug("database trace", sqlLabel, sql, "rows", rows, label.Duration, elapsed)
	}
}

// expand applies GORM's printf-style arguments to message. GORM passes the
// message as a format string; dropping the arguments would leave unexpanded
// verbs in the log record.
func expand(message string, args ...any) string {
	if len(args) == 0 {
		return message
	}
	return fmt.Sprintf(message, args...)
}

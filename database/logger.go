package database

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GormLogger exposes the methods needed for gorm database logging.
type GormLogger interface {
	Print(v ...interface{})
}

// Logger is an adaption of gorm.Logger that uses logrus.
type Logger struct {
	*zap.Logger
}

// GormCallerEncoder finds the file and line number of the first use in gormdb.go.
// The default caller encoder from zapcore is unreliable. Hence this implementation
// ignores the caller argument from zap, and instead we create our own caller for Gorm.
func GormCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	// be careful when modifying these; they have been determined experimentally,
	// but things may change due to changes both internally and externally to Gorm.
	const (
		// lowest is the first stack entry to include in the frame
		lowestEntry = 8
		// highest is the stack entry we need to check
		highestEntry = 28
	)
	pc := make([]uintptr, highestEntry)
	n := runtime.Callers(lowestEntry, pc)
	frames := runtime.CallersFrames(pc[:n])

	var frame runtime.Frame
	more := true
	for more {
		frame, more = frames.Next()
		if strings.HasSuffix(frame.File, "database/gormdb.go") {
			caller = zapcore.EntryCaller{Defined: true, File: frame.File, Line: frame.Line}
			// we may have multiple stack entries from gormdb,
			// but we use the first one; the entry point into gormdb.
			break
		}
	}
	enc.AppendString(caller.TrimmedPath())
}

// NewGormLogger returns a logger for Gorm. A logger instance is only returned
// if the LOGDB environment variable is set.
// This logger should probably not be used in production due to the
// high volume of SQL queries; it is mainly meant to assist with debugging
// database query issues.
func NewGormLogger() GormLogger {
	if os.Getenv("LOGDB") != "" {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeCaller = GormCallerEncoder
		l, _ := cfg.Build()
		return Logger{Logger: l}
	}
	return nil
}

var sqlRegexp = regexp.MustCompile(`(\$\d+)|\?`)

// Print implements the GormLogger interface.
func (l Logger) Print(values ...interface{}) {
	// values[0] = level (sql, log)
	// values[1] = source file:line
	// values[2] = latency (query execution time)
	// values[3] = sql query
	// values[4] = values for query
	// values[5] = affected-rows (if available)
	if len(values) > 1 {
		level := values[0].(string)
		switch level {
		case "sql":
			formattedValues := getFormattedValues(values)
			sql := fmt.Sprintf(sqlRegexp.ReplaceAllString(values[3].(string), "%v"), formattedValues...)
			l := l.With(zap.Any("latency", values[2]))
			if len(values) > 5 {
				l = l.With(zap.Any("affected-rows", values[5]))
			}
			l.Debug(sql)
		default:
			l.Sugar().Info(values[2:]...)
		}
	}
}

func getFormattedValues(values []interface{}) []interface{} {
	rawValues := values[4].([]interface{})
	formattedValues := make([]interface{}, 0, len(rawValues))
	for _, value := range rawValues {
		switch v := value.(type) {
		case time.Time:
			formattedValues = append(formattedValues, fmt.Sprint(v))
		case []byte:
			if str := string(v); isPrintable(str) {
				formattedValues = append(formattedValues, fmt.Sprint(str))
			} else {
				formattedValues = append(formattedValues, "<binary>")
			}
		default:
			str := "NULL"
			if v != nil {
				str = fmt.Sprint(v)
			}
			formattedValues = append(formattedValues, str)
		}
	}
	return formattedValues
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

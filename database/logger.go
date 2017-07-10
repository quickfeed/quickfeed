package database

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"github.com/Sirupsen/logrus"
)

var sqlRegexp = regexp.MustCompile(`(\$\d+)|\?`)

// GormLogger exposes the methods needed for gorm database logging.
type GormLogger interface {
	Print(v ...interface{})
}

// Logger is an adaption of gorm.Logger that uses logrus.
type Logger struct {
	*logrus.Logger
}

// Print implements the GormLogger interface.
func (l Logger) Print(values ...interface{}) {
	if len(values) > 1 {
		level := values[0]
		source := values[1]
		entry := l.WithField("source", source)

		if level == "sql" {
			var formattedValues []interface{}
			for _, value := range values[4].([]interface{}) {
				indirectValue := reflect.Indirect(reflect.ValueOf(value))
				if indirectValue.IsValid() {
					value = indirectValue.Interface()
					if t, ok := value.(time.Time); ok {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format(time.RFC3339)))
					} else if b, ok := value.([]byte); ok {
						if str := string(b); isPrintable(str) {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
						} else {
							formattedValues = append(formattedValues, "'<binary>'")
						}
					} else if r, ok := value.(driver.Valuer); ok {
						if value, err := r.Value(); err == nil && value != nil {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
						} else {
							formattedValues = append(formattedValues, "NULL")
						}
					} else {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
					}
				} else {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
				}
			}

			latency := values[2]
			entry.WithFields(logrus.Fields{
				"time_rfc3339":  time.Now().Format(time.RFC3339),
				"latency_human": latency,
				"latency":       strconv.FormatInt(latency.(time.Duration).Nanoseconds()/1000, 10),
			}).Print(fmt.Sprintf(sqlRegexp.ReplaceAllString(values[3].(string), "%v"), formattedValues...))

		} else {
			l.Error(values[2:]...)
		}
	} else {
		l.Error(values...)
	}
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

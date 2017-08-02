package web

import (
	"io"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
)

// Logger returns a logrus logger middleware.
func Logger(l logrus.FieldLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			r := c.Request()
			w := c.Response()
			start := time.Now()
			if err := next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()

			p := r.URL.Path
			if p == "" {
				p = "/"
			}

			bytesIn := r.Header.Get(echo.HeaderContentLength)
			if bytesIn == "" {
				bytesIn = "0"
			}

			l.WithFields(map[string]interface{}{
				"time_rfc3339":  time.Now().Format(time.RFC3339),
				"remote_ip":     c.RealIP(),
				"host":          r.Host,
				"uri":           r.RequestURI,
				"method":        r.Method,
				"path":          p,
				"referer":       r.Referer(),
				"user_agent":    r.UserAgent(),
				"status":        w.Status,
				"latency":       strconv.FormatInt(stop.Sub(start).Nanoseconds()/1000, 10),
				"latency_human": stop.Sub(start).String(),
				"bytes_in":      bytesIn,
				"bytes_out":     strconv.FormatInt(w.Size, 10),
			}).Info("Handled request")

			return nil
		}
	}
}

// EchoLogger adapts a logrus.Logger to the echo.Logger interface.
type EchoLogger struct {
	*logrus.Logger
}

// Output implements the echo.Logger interface.
func (l EchoLogger) Output() io.Writer {
	return l.Logger.Out
}

// SetOutput implements the echo.Logger interface.
func (l EchoLogger) SetOutput(out io.Writer) {
	l.Logger.Out = out
}

// Prefix implements the echo.Logger interface.
func (l EchoLogger) Prefix() string {
	return ""
}

// SetPrefix implements the echo.Logger interface.
func (l EchoLogger) SetPrefix(p string) {}

// Level implements the echo.Logger interface.
func (l EchoLogger) Level() log.Lvl {
	return log.Lvl(l.Logger.Level)
}

// SetLevel implements the echo.Logger interface.
func (l EchoLogger) SetLevel(level log.Lvl) {
	l.Logger.Level = logrus.Level(level)
}

// Printj implements the echo.Logger interface.
func (l EchoLogger) Printj(json log.JSON) {
	l.Printf("%v", json)
}

// Debugj implements the echo.Logger interface.
func (l EchoLogger) Debugj(json log.JSON) {
	l.Debugf("%v", json)
}

// Infoj implements the echo.Logger interface.
func (l EchoLogger) Infoj(json log.JSON) {
	l.Infof("%v", json)
}

// Warnj implements the echo.Logger interface.
func (l EchoLogger) Warnj(json log.JSON) {
	l.Warnf("%v", json)
}

// Errorj implements the echo.Logger interface.
func (l EchoLogger) Errorj(json log.JSON) {
	l.Errorf("%v", json)
}

// Fatalj implements the echo.Logger interface.
func (l EchoLogger) Fatalj(json log.JSON) {
	l.Fatalf("%v", json)
}

// Panicj implements the echo.Logger interface.
func (l EchoLogger) Panicj(json log.JSON) {
	l.Panicf("%v", json)
}

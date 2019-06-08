package web

import (
	"time"

	"github.com/labstack/echo"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger returns a zap logger middleware.
func Logger(log *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			if err := next(c); err != nil {
				c.Error(err)
			}
			latency := time.Since(start)

			req := c.Request()
			res := c.Response()

			p := req.URL.Path
			if p == "" {
				p = "/"
			}

			bytesIn := req.Header.Get(echo.HeaderContentLength)
			if bytesIn == "" {
				bytesIn = "0"
			}

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}

			fields := []zapcore.Field{
				zap.Int("status", res.Status),
				zap.Duration("latency", latency),
				zap.String("id", id),
				zap.String("method", req.Method),
				zap.String("uri", req.RequestURI),
				zap.String("host", req.Host),
				zap.String("remote_ip", c.RealIP()),
				zap.String("path", p),
				zap.String("referer", req.Referer()),
				zap.String("user_agent", req.UserAgent()),
				zap.String("bytes_in", bytesIn),
				zap.Int64("bytes_out", res.Size),
			}

			n := res.Status
			switch {
			case n >= 500:
				log.Error("Server error", fields...)
			case n >= 400:
				log.Warn("Client error", fields...)
			case n >= 300:
				log.Info("Redirection", fields...)
			default:
				log.Info("Success", fields...)
			}
			return nil
		}
	}
}

// WebhookLogger implements the gopkg.in/go-playground/webhooks.v3.Logger
// interface using a zap.Logger.
//TODO(meling) with webhooks.v5 it does not use a logger; so we should not need this?
type WebhookLogger struct {
	*zap.Logger
}

// Info prints basic information.
func (l WebhookLogger) Info(msg string) {
	l.Info(msg)
}

// Error prints error information.
func (l WebhookLogger) Error(msg string) {
	l.Error(msg)
}

// Debug prints information useful for debugging.
func (l WebhookLogger) Debug(msg string) {
	l.Debug(msg)
}

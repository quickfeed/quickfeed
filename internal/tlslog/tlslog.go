// Package tlslog turns the TLS handshake failures that net/http writes to
// http.Server.ErrorLog into counters, sampled records, and periodic summaries.
//
// A public server sees a steady trickle of failed handshakes from scanners,
// from clients using obsolete TLS settings, and from connections carrying an
// arbitrary server name. These failures happen before any request reaches
// QuickFeed and are not actionable; logging one error per connection hides the
// server errors that are. New returns a *log.Logger for http.Server.ErrorLog
// that counts each expected failure, keeps one representative record per
// reason per interval, and summarizes the rest. Handshake failures that match
// no expected reason, and every message that is not a handshake failure at
// all, are passed through so that accept failures, handler panics and similar
// server errors stay visible.
package tlslog

import (
	"fmt"
	"log"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/quickfeed/quickfeed/internal/qlog/label"
)

// prefix is the fixed text that net/http puts in front of every TLS handshake
// failure it writes to http.Server.ErrorLog; see net/http.conn.serve.
const prefix = "http: TLS handshake error from "

// component is the value of the label.Component attribute on the records this
// package emits, so that handshake noise can be filtered out of the log.
const component = "tls"

// reasonKey is the attribute and Prometheus label naming the failure class. It
// is local to this package because nothing else logs or queries it.
const reasonKey = "reason"

// The fixed set of reasons a handshake failure is classified into. The set is
// deliberately small and closed, because it is used as a Prometheus label
// value: client addresses, ports and requested host names must never become
// label values, since they are unbounded.
const (
	reasonUnauthorizedSNI = "unauthorized_sni"
	reasonProtocolProbe   = "protocol_probe"
	reasonDisconnected    = "disconnected"
	reasonTimeout         = "timeout"
	reasonUnexpected      = "unexpected"
)

// reasons is every reason, with unexpected last.
var reasons = []string{
	reasonUnauthorizedSNI,
	reasonProtocolProbe,
	reasonDisconnected,
	reasonTimeout,
	reasonUnexpected,
}

// expectedReasons are the reasons that are counted and sampled, in the order
// they appear in a summary. unexpected is not among them, since such failures
// are logged in full instead.
var expectedReasons = reasons[:len(reasons)-1]

// DefaultInterval is how long a sampled record and the counts behind a summary
// cover, unless WithInterval says otherwise.
const DefaultInterval = 10 * time.Minute

// Option configures the error logger returned by New.
type Option func(*sampler)

// WithLogger routes the classified output to logger instead of the default
// logger. Without it, records go to whichever logger is installed as the slog
// default when the failure occurs.
func WithLogger(logger *slog.Logger) Option {
	return func(s *sampler) { s.target = logger }
}

// WithInterval sets how often a reason may produce a sampled record, and how
// long a summary covers.
func WithInterval(interval time.Duration) Option {
	return func(s *sampler) { s.interval = interval }
}

// WithClock replaces the clock used for sampling and summaries, so that tests
// can drive time.
func WithClock(now func() time.Time) Option {
	return func(s *sampler) { s.now = now }
}

// New returns a logger to assign to http.Server.ErrorLog. The logger counts
// and classifies TLS handshake failures, and writes what survives sampling to
// the QuickFeed logger.
func New(opts ...Option) *log.Logger {
	return log.New(newSampler(opts...), "", 0)
}

// sampler is the io.Writer behind the returned log.Logger. It holds the counts
// of the current summary interval and when each reason was last sampled.
type sampler struct {
	target   *slog.Logger
	now      func() time.Time
	interval time.Duration

	mu          sync.Mutex
	counts      map[string]int
	lastSampled map[string]time.Time
	windowStart time.Time // first failure of the current summary interval; zero when no failure has occurred since the last summary.
}

func newSampler(opts ...Option) *sampler {
	s := &sampler{
		now:         time.Now,
		interval:    DefaultInterval,
		counts:      make(map[string]int),
		lastSampled: make(map[string]time.Time),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Write implements io.Writer for the log.Logger returned by New. The standard
// logger passes one complete message per call, with a trailing newline.
func (s *sampler) Write(p []byte) (int, error) {
	s.record(strings.TrimRight(string(p), "\n"))
	return len(p), nil
}

func (s *sampler) record(msg string) {
	logger := s.logger()
	reason, ok := classify(msg)
	if !ok {
		// Not a handshake failure: an accept failure, a handler panic, or
		// another server error that must reach the log unchanged.
		logger.Error(msg)
		return
	}
	handshakeFailures.WithLabelValues(reason).Inc()
	summary, sample := s.update(reason)
	if summary != "" {
		logger.Info(summary, label.Component, component)
	}
	if reason == reasonUnexpected {
		// An unclassified handshake failure may be actionable, so it keeps its
		// original message and is never sampled away.
		logger.Warn(msg, label.Component, component, reasonKey, reason)
		return
	}
	if sample {
		logger.Info(msg, label.Component, component, reasonKey, reason)
	}
}

// update records one failure. It returns the summary of the preceding interval
// if that interval has elapsed, and whether this failure is the representative
// sample for its reason. Summaries are produced here, rather than by a ticker,
// so that the sampler owns no goroutine: a summary describes failures, so the
// next failure is a fine time to emit it.
func (s *sampler) update(reason string) (summary string, sample bool) {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.windowStart.IsZero() && now.Sub(s.windowStart) >= s.interval {
		summary = s.summary()
		clear(s.counts)
		s.windowStart = time.Time{}
	}
	if reason == reasonUnexpected {
		return summary, false
	}
	if s.windowStart.IsZero() {
		s.windowStart = now
	}
	s.counts[reason]++
	if last, seen := s.lastSampled[reason]; !seen || now.Sub(last) >= s.interval {
		s.lastSampled[reason] = now
		sample = true
	}
	return summary, sample
}

// summary renders the counts of the elapsed interval. The caller holds s.mu.
func (s *sampler) summary() string {
	var b strings.Builder
	fmt.Fprintf(&b, "TLS handshake failures in last %s:", shortDuration(s.interval))
	for _, reason := range expectedReasons {
		if count := s.counts[reason]; count > 0 {
			fmt.Fprintf(&b, " %s=%d", reason, count)
		}
	}
	return b.String()
}

func (s *sampler) logger() *slog.Logger {
	if s.target != nil {
		return s.target
	}
	return slog.Default()
}

// classify returns the reason for the TLS handshake failure described by msg,
// and reports whether msg is a handshake failure at all.
func classify(msg string) (string, bool) {
	detail, ok := failureDetail(msg)
	if !ok {
		return "", false
	}
	switch {
	// Checked first, since the rejected host name is attacker-chosen and could
	// otherwise contain the text another reason matches on.
	case containsAny(detail, "not configured in HostWhitelist", "missing server name"):
		return reasonUnauthorizedSNI, true
	case containsAny(detail,
		"unsupported application protocols",
		"unsupported versions",
		"no cipher suite supported",
		"sent an HTTP request to an HTTPS server",
		"does not look like a TLS handshake",
	):
		return reasonProtocolProbe, true
	case strings.Contains(detail, "timeout"):
		return reasonTimeout, true
	case strings.HasSuffix(detail, "EOF"),
		containsAny(detail, "connection reset by peer", "broken pipe"):
		return reasonDisconnected, true
	}
	return reasonUnexpected, true
}

// failureDetail returns the part of a handshake failure message that describes
// the failure, that is, the message without its prefix and client address, and
// reports whether msg is a handshake failure. Dropping the address keeps the
// classification from matching on an address or port number.
func failureDetail(msg string) (string, bool) {
	rest, ok := strings.CutPrefix(msg, prefix)
	if !ok {
		return "", false
	}
	if _, detail, ok := strings.Cut(rest, ": "); ok {
		return detail, true
	}
	return rest, true
}

func containsAny(s string, substrings ...string) bool {
	for _, substring := range substrings {
		if strings.Contains(s, substring) {
			return true
		}
	}
	return false
}

// shortDuration renders a whole number of hours, minutes or seconds without
// the zero-valued units that time.Duration's own String method keeps.
func shortDuration(d time.Duration) string {
	switch {
	case d >= time.Hour && d%time.Hour == 0:
		return fmt.Sprintf("%dh", d/time.Hour)
	case d >= time.Minute && d%time.Minute == 0:
		return fmt.Sprintf("%dm", d/time.Minute)
	case d >= time.Second && d%time.Second == 0:
		return fmt.Sprintf("%ds", d/time.Second)
	}
	return d.String()
}

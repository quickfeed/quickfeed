package tlslog

import (
	"context"
	"log"
	"log/slog"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/quickfeed/quickfeed/internal/qlog/label"
)

// The sample messages below are the ones recorded in the issue that motivated
// this package; the client addresses have been replaced.
const (
	rejectedHost    = prefix + `192.0.2.10:51514: acme/autocert: host "crm.itest.run" not configured in HostWhitelist`
	rejectedHost2   = prefix + `192.0.2.11:44012: acme/autocert: host "pedersen.itest.run" not configured in HostWhitelist`
	missingHost     = prefix + "192.0.2.12:33218: acme/autocert: missing server name"
	unsupportedALPN = prefix + `192.0.2.13:57002: tls: client requested unsupported application protocols (["http/0.9" "http/1.0" "spdy/1" "spdy/2" "spdy/3" "h2c" "hq"])`
	unsupportedTLS  = prefix + "192.0.2.14:41118: tls: client offered only unsupported versions: [302 301]"
	noCipherSuite   = prefix + "192.0.2.15:60122: tls: no cipher suite supported by both client and server"
	plaintextHTTP   = prefix + "192.0.2.16:38810: client sent an HTTP request to an HTTPS server"
	notTLS          = prefix + "192.0.2.17:49930: tls: first record does not look like a TLS handshake"
	eof             = prefix + "192.0.2.18:52144: EOF"
	connectionReset = prefix + "192.0.2.19:40088: read: connection reset by peer"
	ioTimeout       = prefix + "192.0.2.20:35566: i/o timeout"
	unknownFailure  = prefix + "192.0.2.21:47722: tls: internal error: unexplained failure"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name          string
		msg           string
		wantReason    string
		wantHandshake bool
	}{
		{name: "rejected host", msg: rejectedHost, wantReason: reasonUnauthorizedSNI, wantHandshake: true},
		{name: "rejected host 2", msg: rejectedHost2, wantReason: reasonUnauthorizedSNI, wantHandshake: true},
		{name: "missing server name", msg: missingHost, wantReason: reasonUnauthorizedSNI, wantHandshake: true},
		{name: "unsupported application protocols", msg: unsupportedALPN, wantReason: reasonProtocolProbe, wantHandshake: true},
		{name: "unsupported TLS versions", msg: unsupportedTLS, wantReason: reasonProtocolProbe, wantHandshake: true},
		{name: "no cipher suite", msg: noCipherSuite, wantReason: reasonProtocolProbe, wantHandshake: true},
		{name: "plaintext HTTP", msg: plaintextHTTP, wantReason: reasonProtocolProbe, wantHandshake: true},
		{name: "not a TLS record", msg: notTLS, wantReason: reasonProtocolProbe, wantHandshake: true},
		{name: "EOF", msg: eof, wantReason: reasonDisconnected, wantHandshake: true},
		{name: "connection reset", msg: connectionReset, wantReason: reasonDisconnected, wantHandshake: true},
		{name: "i/o timeout", msg: ioTimeout, wantReason: reasonTimeout, wantHandshake: true},
		{name: "unknown failure", msg: unknownFailure, wantReason: reasonUnexpected, wantHandshake: true},
		{
			// The rejected host name comes from the client, so it must not be
			// able to steer the classification towards another reason.
			name:          "rejected host named after another reason",
			msg:           prefix + `192.0.2.22:41000: acme/autocert: host "i/o timeout.itest.run" not configured in HostWhitelist`,
			wantReason:    reasonUnauthorizedSNI,
			wantHandshake: true,
		},
		{
			name:          "IPv6 client",
			msg:           prefix + "[2001:db8::1]:51514: EOF",
			wantReason:    reasonDisconnected,
			wantHandshake: true,
		},
		{name: "accept failure", msg: "http: Accept error: accept tcp [::]:443: too many open files; retrying in 5ms", wantHandshake: false},
		{name: "handler panic", msg: "http: panic serving 192.0.2.23:1234: runtime error: invalid memory address", wantHandshake: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason, isHandshake := classify(tt.msg)
			if isHandshake != tt.wantHandshake {
				t.Fatalf("classify(%q) handshake = %t, want %t", tt.msg, isHandshake, tt.wantHandshake)
			}
			if reason != tt.wantReason {
				t.Errorf("classify(%q) = %q, want %q", tt.msg, reason, tt.wantReason)
			}
		})
	}
}

// TestPassThrough checks that a message that is not a TLS handshake failure
// reaches the log unchanged, at error level.
func TestPassThrough(t *testing.T) {
	recorder, errorLog, _ := newTestLogger(t)
	const msg = "http: Accept error: accept tcp [::]:443: too many open files; retrying in 5ms"
	errorLog.Print(msg)

	recorder.check(t, []record{{level: slog.LevelError, msg: msg}})
}

// TestSample checks that the first failure of a reason is logged in full, and
// that the failures following it within the interval are not.
func TestSample(t *testing.T) {
	recorder, errorLog, clock := newTestLogger(t)
	before := counted(t, reasonUnauthorizedSNI)

	errorLog.Print(rejectedHost)
	clock.advance(time.Minute)
	errorLog.Print(rejectedHost2)
	clock.advance(time.Minute)
	errorLog.Print(missingHost)

	recorder.check(t, []record{{
		level:  slog.LevelInfo,
		msg:    rejectedHost,
		reason: reasonUnauthorizedSNI,
	}})
	if got := counted(t, reasonUnauthorizedSNI) - before; got != 3 {
		t.Errorf("counted %v failures, want 3", got)
	}
}

// TestSampleEachReason checks that the reasons are sampled independently of
// each other.
func TestSampleEachReason(t *testing.T) {
	recorder, errorLog, clock := newTestLogger(t)

	errorLog.Print(rejectedHost)
	clock.advance(time.Second)
	errorLog.Print(noCipherSuite)
	clock.advance(time.Second)
	errorLog.Print(rejectedHost2)

	recorder.check(t, []record{
		{level: slog.LevelInfo, msg: rejectedHost, reason: reasonUnauthorizedSNI},
		{level: slog.LevelInfo, msg: noCipherSuite, reason: reasonProtocolProbe},
	})
}

// TestSummary checks that the failures suppressed by sampling are summarized
// once the interval has elapsed, and that the sampling then starts over.
func TestSummary(t *testing.T) {
	recorder, errorLog, clock := newTestLogger(t)

	errorLog.Print(rejectedHost)  // sampled
	errorLog.Print(rejectedHost2) // counted only
	errorLog.Print(missingHost)   // counted only
	errorLog.Print(noCipherSuite) // sampled
	errorLog.Print(unsupportedTLS)
	errorLog.Print(eof)
	errorLog.Print(ioTimeout)
	clock.advance(testInterval)
	errorLog.Print(rejectedHost2)

	recorder.check(t, []record{
		{level: slog.LevelInfo, msg: rejectedHost, reason: reasonUnauthorizedSNI},
		{level: slog.LevelInfo, msg: noCipherSuite, reason: reasonProtocolProbe},
		{level: slog.LevelInfo, msg: eof, reason: reasonDisconnected},
		{level: slog.LevelInfo, msg: ioTimeout, reason: reasonTimeout},
		{
			level: slog.LevelInfo,
			msg:   "TLS handshake failures in last 10m: unauthorized_sni=3 protocol_probe=2 disconnected=1 timeout=1",
		},
		// The interval has elapsed, so this failure is sampled again.
		{level: slog.LevelInfo, msg: rejectedHost2, reason: reasonUnauthorizedSNI},
	})
}

// TestSummaryCoversOnlyElapsedInterval checks that a summary reports the
// failures of the interval that just ended, and not those of earlier ones.
func TestSummaryCoversOnlyElapsedInterval(t *testing.T) {
	recorder, errorLog, clock := newTestLogger(t)

	errorLog.Print(eof)
	errorLog.Print(eof)
	clock.advance(testInterval)
	errorLog.Print(eof)
	clock.advance(testInterval)
	errorLog.Print(connectionReset)

	recorder.check(t, []record{
		{level: slog.LevelInfo, msg: eof, reason: reasonDisconnected},
		{level: slog.LevelInfo, msg: "TLS handshake failures in last 10m: disconnected=2"},
		{level: slog.LevelInfo, msg: eof, reason: reasonDisconnected},
		{level: slog.LevelInfo, msg: "TLS handshake failures in last 10m: disconnected=1"},
		{level: slog.LevelInfo, msg: connectionReset, reason: reasonDisconnected},
	})
}

// TestUnexpected checks that a handshake failure that matches no expected
// reason keeps its full message, is never sampled away, and stays out of the
// summary counts.
func TestUnexpected(t *testing.T) {
	recorder, errorLog, clock := newTestLogger(t)
	before := counted(t, reasonUnexpected)

	errorLog.Print(eof)
	errorLog.Print(unknownFailure)
	errorLog.Print(unknownFailure)
	clock.advance(testInterval)
	errorLog.Print(unknownFailure)

	recorder.check(t, []record{
		{level: slog.LevelInfo, msg: eof, reason: reasonDisconnected},
		{level: slog.LevelWarn, msg: unknownFailure, reason: reasonUnexpected},
		{level: slog.LevelWarn, msg: unknownFailure, reason: reasonUnexpected},
		{level: slog.LevelInfo, msg: "TLS handshake failures in last 10m: disconnected=1"},
		{level: slog.LevelWarn, msg: unknownFailure, reason: reasonUnexpected},
	})
	if got := counted(t, reasonUnexpected) - before; got != 3 {
		t.Errorf("counted %v failures, want 3", got)
	}
}

func TestShortDuration(t *testing.T) {
	tests := []struct {
		interval time.Duration
		want     string
	}{
		{interval: 10 * time.Minute, want: "10m"},
		{interval: time.Hour, want: "1h"},
		{interval: 30 * time.Second, want: "30s"},
		{interval: 90 * time.Second, want: "90s"},
		{interval: 1500 * time.Millisecond, want: "1.5s"},
	}
	for _, tt := range tests {
		if got := shortDuration(tt.interval); got != tt.want {
			t.Errorf("shortDuration(%s) = %q, want %q", tt.interval, got, tt.want)
		}
	}
}

const testInterval = 10 * time.Minute

// newTestLogger returns an error logger of the kind installed on
// http.Server.ErrorLog, along with the records it produces and the clock that
// drives its sampling and summaries.
func newTestLogger(t *testing.T) (*recorder, *log.Logger, *testClock) {
	t.Helper()
	rec := &recorder{}
	clock := &testClock{now: time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)}
	errorLog := New(
		WithLogger(slog.New(rec)),
		WithInterval(testInterval),
		WithClock(clock.Now),
	)
	return rec, errorLog, clock
}

// counted returns the current value of the handshake failure counter for
// reason. The counter is process-wide, so tests compare differences.
func counted(t *testing.T, reason string) float64 {
	t.Helper()
	return testutil.ToFloat64(handshakeFailures.WithLabelValues(reason))
}

type testClock struct {
	now time.Time
}

func (c *testClock) Now() time.Time { return c.now }

func (c *testClock) advance(d time.Duration) { c.now = c.now.Add(d) }

// record is the part of a log record the tests assert on.
type record struct {
	level     slog.Level
	msg       string
	reason    string // expected value of the reason attribute; empty if the record must not carry one.
	component string
}

// recorder is a slog.Handler that keeps the records it is given.
type recorder struct {
	records []record
}

func (r *recorder) Enabled(context.Context, slog.Level) bool { return true }

func (r *recorder) Handle(_ context.Context, rec slog.Record) error {
	entry := record{level: rec.Level, msg: rec.Message}
	rec.Attrs(func(attr slog.Attr) bool {
		switch attr.Key {
		case reasonKey:
			entry.reason = attr.Value.String()
		case label.Component:
			entry.component = attr.Value.String()
		}
		return true
	})
	r.records = append(r.records, entry)
	return nil
}

func (r *recorder) WithAttrs([]slog.Attr) slog.Handler { return r }

func (r *recorder) WithGroup(string) slog.Handler { return r }

// check compares the recorded records with want. Every record except a
// pass-through one, the only kind logged at error level, carries the component
// attribute, so want need not repeat it.
func (r *recorder) check(t *testing.T, want []record) {
	t.Helper()
	for i := range want {
		if want[i].level != slog.LevelError {
			want[i].component = component
		}
	}
	if len(r.records) != len(want) {
		t.Fatalf("logged %d records, want %d:\ngot:  %+v\nwant: %+v", len(r.records), len(want), r.records, want)
	}
	for i, got := range r.records {
		if got != want[i] {
			t.Errorf("record %d = %+v, want %+v", i, got, want[i])
		}
	}
}

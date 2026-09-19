package tlslog

import "github.com/prometheus/client_golang/prometheus"

// Collectors returns the Prometheus collectors for TLS handshake failures, for
// registration by the metrics package.
func Collectors() []prometheus.Collector {
	return []prometheus.Collector{handshakeFailures}
}

var handshakeFailures = newHandshakeFailures()

func newHandshakeFailures() *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "quickfeed_tls_handshake_failures_total",
		Help: "Total number of failed TLS handshakes, by classified reason.",
	}, []string{reasonKey})
	// Create every series up front, so that a query for a reason that has not
	// occurred yet returns zero rather than no series at all.
	for _, reason := range reasons {
		counter.WithLabelValues(reason)
	}
	return counter
}

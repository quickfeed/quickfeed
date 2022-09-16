package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/web/interceptor"
)

var reg = prometheus.NewRegistry()

func init() {
	metricsCollectorsSets := [][]prometheus.Collector{
		interceptor.RPCMetricsCollectors(),
		ci.TestExecutionMetricsCollectors(),
	}
	for _, collectors := range metricsCollectorsSets {
		reg.MustRegister(collectors...)
	}
}

func Handler() http.Handler {
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}

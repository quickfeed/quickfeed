package ag

import "github.com/prometheus/client_golang/prometheus"

var (
	// AgResponseTimeByMethodsMetric records response time by method name
	AgResponseTimeByMethodsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ag_response_time_by_method",
	}, []string{"method"})

	// AgFailedMethodsMetric counts amount of times every method resulted in error
	AgFailedMethodsMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ag_methods_failed",
	}, []string{"method"})

	// AgResponsePayloadSizeMetric records response size in bytes for every method
	AgResponsePayloadSizeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ag_payload_size_by_method",
	}, []string{"method"})
)

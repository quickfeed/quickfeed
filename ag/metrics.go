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

	// AgMethodSuccessRateMetric counts the amount of calls for every method, allows
	// grouping by method name and by result ("total", "success", "error")
	AgMethodSuccessRateMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ag_method_success_failure_rate",
	}, []string{"method", "result"})
)

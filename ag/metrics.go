package ag

import "github.com/prometheus/client_golang/prometheus"

var (
	// CustomizedCounterMetric creates a customized prometheus counter metric.
	CustomizedCounterMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ag_server_grpc_method_calls_handle_count",
		Help: "Total number of RPCs handled on the server.",
	}, []string{"name"})

	// CustomizedResponseTimeMetric describes response time for grpc methods
	CustomizedResponseTimeMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ag_server_method_response_time",
	}, []string{"name"})

	// GetCourseLabSubmissionsMetric describes response times for the slowest method
	GetCourseLabSubmissionsMetric = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "ag_server_GetCourseLabSubmissions_response_time",
	})
)

package ci

import "github.com/prometheus/client_golang/prometheus"

// TestExecutionMetricsCollectors returns a list of Prometheus metrics collectors for test execution.
func TestExecutionMetricsCollectors() []prometheus.Collector {
	return []prometheus.Collector{
		cloneTimeGauge,
		validationTimeGauge,
		testExecutionTimeGauge,
		testsStartedCounter,
		testsFailedCounter,
		testsSucceededCounter,
	}
}

var (
	cloneTimeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "quickfeed_clone_repositories_time",
		Help: "The time to clone tests and student repository for test execution.",
	}, []string{"user", "course"})

	validationTimeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "quickfeed_repository_validation_time",
		Help: "The time to validate student repository for issues.",
	}, []string{"user", "course"})

	testExecutionTimeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "quickfeed_test_execution_time",
		Help: "The time to run test execution.",
	}, []string{"user", "course"})

	testsStartedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "quickfeed_test_execution_attempts",
		Help: "Total number of times test execution was attempted",
	}, []string{"user", "course"})

	testsFailedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "quickfeed_test_execution_failed",
		Help: "Total number of times test execution failed",
	}, []string{"user", "course"})

	testsFailedWithOutputCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "quickfeed_test_execution_failed_with_output",
		Help: "Total number of times test execution failed with output",
	}, []string{"user", "course"})

	testsFailedExtractResultsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "quickfeed_test_execution_failed_to_extract_results",
		Help: "Total number of times test execution failed to extract results",
	}, []string{"user", "course"})

	testsSucceededCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "quickfeed_test_execution_succeeded",
		Help: "Total number of times test execution succeeded",
	}, []string{"user", "course"})
)

func timer(jobOwner, course string, gauge *prometheus.GaugeVec) func() {
	responseTimer := prometheus.NewTimer(prometheus.ObserverFunc(
		gauge.WithLabelValues(jobOwner, course).Set),
	)
	return func() { responseTimer.ObserveDuration() }
}

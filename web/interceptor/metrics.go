package interceptor

import (
	"context"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/prometheus/client_golang/prometheus"
)

// RPCMetricsCollectors returns a list of Prometheus metrics collectors for RPC related metrics.
func RPCMetricsCollectors() []prometheus.Collector {
	return []prometheus.Collector{
		loginCounter,
		failedMethodsCounter,
		accessedMethodsCounter,
		respondedMethodsCounter,
		responseTimeGauge,
	}
}

var (
	responseTimeGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "quickfeed_method_response_time",
		Help: "The response time for method.",
	}, []string{"method"})

	accessedMethodsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "quickfeed_method_accessed",
		Help: "Total number of times method accessed",
	}, []string{"method"})

	respondedMethodsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "quickfeed_method_responded",
		Help: "Total number of times method responded successfully",
	}, []string{"method"})

	failedMethodsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "quickfeed_method_failed",
		Help: "Total number of times method failed with an error",
	}, []string{"method"})

	loginCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "quickfeed_login_attempts",
		Help: "Total number of login attempts",
	}, []string{"user"})
)

func Metrics() connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			procedure := request.Spec().Procedure
			methodName := procedure[strings.LastIndex(procedure, "/")+1:]
			defer metricsTimer(methodName)()
			resp, err := next(ctx, request)
			accessedMethodsCounter.WithLabelValues(methodName).Inc()
			if resp != nil {
				respondedMethodsCounter.WithLabelValues(methodName).Inc()
			}
			if err != nil {
				failedMethodsCounter.WithLabelValues(methodName).Inc()
				if methodName == "GetUser" {
					// Can't get the user ID from err; so just counting
					loginCounter.WithLabelValues("").Inc()
				}
			}
			return resp, err
		})
	})
}

func metricsTimer(methodName string) func() {
	responseTimer := prometheus.NewTimer(prometheus.ObserverFunc(
		responseTimeGauge.WithLabelValues(methodName).Set),
	)
	return func() { responseTimer.ObserveDuration() }
}

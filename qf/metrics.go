package qf

import (
	context "context"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
)

var (
	// AgResponseTimeByMethodsMetric records response time by method name
	AgResponseTimeByMethodsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ag_response_time",
	}, []string{"method"})

	// AgFailedMethodsMetric counts amount of times every method resulted in error
	AgFailedMethodsMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ag_methods_failed",
	}, []string{"method"})

	// AgMethodSuccessRateMetric counts the amount of calls for every method, allows
	// grouping by method name and by result ("total", "success", "error")
	AgMethodSuccessRateMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ag_success_rate",
	}, []string{"method", "result"})
)

func MetricsInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		methodName := info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:]
		defer metricsTimer(methodName)()
		resp, err := handler(ctx, req)
		handleMetrics(methodName, resp, err)
		return resp, err
	}
}

func metricsTimer(methodName string) func() {
	responseTimer := prometheus.NewTimer(prometheus.ObserverFunc(
		AgResponseTimeByMethodsMetric.WithLabelValues(methodName).Set),
	)
	return func() {
		responseTimer.ObserveDuration().Milliseconds()
	}
}

func handleMetrics(methodName string, resp interface{}, err error) {
	AgMethodSuccessRateMetric.WithLabelValues(methodName, "total").Inc()
	if resp != nil {
		AgMethodSuccessRateMetric.WithLabelValues(methodName, "success").Inc()
	}
	if err != nil {
		AgFailedMethodsMetric.WithLabelValues(methodName).Inc()
		AgMethodSuccessRateMetric.WithLabelValues(methodName, "error").Inc()
	}
}

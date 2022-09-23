package interceptor

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bufbuild/connect-go"
	promgrpc "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Create a metrics registry.
var reg = prometheus.NewRegistry()

func init() {
	reg.MustRegister(
		promgrpc.NewServerMetrics(),
		FailedMethodsMetric,
		MethodSuccessRateMetric,
		ResponseTimeByMethodsMetric,
	)
}

// MetricsServer returns a HTTP Server that serves the prometheus metrics.
func MetricsServer(port int) *http.Server {
	return &http.Server{
		Handler:           promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
		Addr:              fmt.Sprintf("127.0.0.1:%d", port),
		ReadHeaderTimeout: 3 * time.Second, // to prevent Slowloris (CWE-400)
	}
}

var (
	// ResponseTimeByMethodsMetric records response time by method name.
	ResponseTimeByMethodsMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "response_time",
	}, []string{"method"})

	// FailedMethodsMetric counts the number of times every method resulted in error.
	FailedMethodsMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "methods_failed",
	}, []string{"method"})

	// MethodSuccessRateMetric counts the number of calls for every method, allows
	// grouping by method name and by result ("total", "success", "error")
	MethodSuccessRateMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "success_rate",
	}, []string{"method", "result"})
)

type MetricsInterceptor struct {
}

func NewMetricsInterceptor() *MetricsInterceptor {
	return &MetricsInterceptor{}
}

func (*MetricsInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		procedure := conn.Spec().Procedure
		methodName := procedure[strings.LastIndex(procedure, "/")+1:]
		defer metricsTimer(methodName)()
		err := next(ctx, conn)
		handleMetrics(methodName, nil, err)
		return err
	})
}

func (*MetricsInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec)
	})
}

func (*MetricsInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		procedure := request.Spec().Procedure
		methodName := procedure[strings.LastIndex(procedure, "/")+1:]
		defer metricsTimer(methodName)()
		resp, err := next(ctx, request)
		handleMetrics(methodName, resp, err)
		return resp, err
	})
}

func metricsTimer(methodName string) func() {
	responseTimer := prometheus.NewTimer(prometheus.ObserverFunc(
		ResponseTimeByMethodsMetric.WithLabelValues(methodName).Set),
	)
	return func() {
		responseTimer.ObserveDuration()
	}
}

func handleMetrics(methodName string, resp interface{}, err error) {
	MethodSuccessRateMetric.WithLabelValues(methodName, "total").Inc()
	if resp != nil {
		MethodSuccessRateMetric.WithLabelValues(methodName, "success").Inc()
	}
	if err != nil {
		FailedMethodsMetric.WithLabelValues(methodName).Inc()
		MethodSuccessRateMetric.WithLabelValues(methodName, "error").Inc()
	}
}

package interceptors

import (
	"context"
	"reflect"
	"strings"
	"time"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MaxWait is the maximum time a request is allowed to stay open before aborting.
const MaxWait = 2 * time.Minute

// TODO(vera): this constand does not belong here,
// but putting it in config package causes import cycle

type validator interface {
	IsValid() bool
}

type idCleaner interface {
	RemoveRemoteID()
}

// ValidateMethod returns a new unary server interceptor that validates requests
// that implements the validator interface.
// Invalid requests are rejected without logging and before it reaches any
// user-level code and returns an illegal argument to the client.
// In addition, the interceptor also implements a cancel mechanism.
func ValidateMethod(logger *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger.Debug("VALIDATE INTERCEPTOR")
		methodName := info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:]
		pb.AgMethodSuccessRateMetric.WithLabelValues(methodName, "total").Inc()
		responseTimer := prometheus.NewTimer(prometheus.ObserverFunc(
			pb.AgResponseTimeByMethodsMetric.WithLabelValues(methodName).Set),
		)
		defer responseTimer.ObserveDuration().Milliseconds()

		if v, ok := req.(validator); ok {
			if !v.IsValid() {
				return nil, status.Errorf(codes.InvalidArgument, "invalid payload")
			}
		} else {
			// just logging, but still handling the call
			logger.Debugf("message type '%s' does not implement validator interface",
				reflect.TypeOf(req).String())
		}
		ctx, cancel := context.WithTimeout(ctx, MaxWait)
		defer cancel()

		// if response has information on remote ID, it will be removed
		resp, err := handler(ctx, req)
		if resp != nil {
			pb.AgMethodSuccessRateMetric.WithLabelValues(methodName, "success").Inc()
			if v, ok := resp.(idCleaner); ok {
				v.RemoveRemoteID()
			}
		}
		if err != nil {
			pb.AgFailedMethodsMetric.WithLabelValues(methodName).Inc()
			pb.AgMethodSuccessRateMetric.WithLabelValues(methodName, "error").Inc()
		}
		return resp, err
	}
}

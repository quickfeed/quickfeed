package config

import (
	"net/http"

	"github.com/autograde/quickfeed/database"
	"github.com/autograde/quickfeed/web/auth"
	"github.com/autograde/quickfeed/web/auth/interceptors"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GrpcMultiplexer struct {
	*grpcweb.WrappedGrpcServer
}

// GenerateTLSApi will load TLS certificates and key and create a grpc server with those.
func (conf *Config) GenerateTLSApi(logger *zap.SugaredLogger, db database.Database, tokens *auth.TokenManager) (*grpc.Server, error) {
	cred, err := credentials.NewServerTLSFromFile(conf.Paths.CertPath, conf.Paths.CertKeyPath)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer(
		grpc.Creds(cred),
		grpc.ChainUnaryInterceptor(
			interceptors.ValidateRequest(logger),
			interceptors.ValidateToken(logger, tokens),
			interceptors.UpdateTokens(logger, tokens),
			interceptors.AccessControl(logger, db, tokens),
		),
	)
	return s, nil
}

// MultiplexHandler is used to route requests to either grpc or to regular http
func (m *GrpcMultiplexer) MultiplexerHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.IsGrpcWebRequest(r) {
			m.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

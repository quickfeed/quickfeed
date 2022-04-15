package config

import (
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GrpcMultiplexer struct {
	Server *grpcweb.WrappedGrpcServer
}

// GenerateTLSApi will load TLS certificates and key and create a grpc server with those.
func (conf *Config) GenerateTLSApi() (*grpc.Server, error) {
	cred, err := credentials.NewServerTLSFromFile(conf.Paths.PemPath, conf.Paths.KeyPath)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer(
		grpc.Creds(cred),
		grpc.ChainUnaryInterceptor(
		// interceptors.ValidateRequest,
		// interceptors.ValidateUser,
		// interceptors.AccessControl,
		),
	)
	return s, nil
}

// MultiplexHandler is used to route requests to either grpc or to regular http
func (m *GrpcMultiplexer) MultiplexerHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.Server.IsGrpcWebRequest(r) {
			m.Server.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

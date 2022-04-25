package config

import (
	"log"
	"net/http"

	"github.com/autograde/quickfeed/web/auth"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GrpcMultiplexer struct {
	*grpcweb.WrappedGrpcServer
}

// GenerateTLSApi will load TLS certificates and key and create a grpc server with those.
func (conf *Config) GenerateTLSApi(logger *zap.Logger, tokens *auth.TokenManager) (*grpc.Server, error) {
	cred, err := credentials.NewServerTLSFromFile(conf.Paths.PemPath, conf.Paths.KeyPath)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer(
		grpc.Creds(cred),
		// grpc.ChainUnaryInterceptor(
		// 	interceptors.ValidateMethod(logger),
		// 	interceptors.UpdateTokens(logger, tokens),
		// 	interceptors.ValidateToken(logger, tokens),
		// 	interceptors.AccessControl(logger, tokens),
		//),
	)
	return s, nil
}

// MultiplexHandler is used to route requests to either grpc or to regular http
func (m *GrpcMultiplexer) MultiplexerHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("MULTIPLEX: %+v", r.RequestURI)
		if m.IsGrpcWebRequest(r) {
			log.Println("MULTIPLEX: grpc. ", r.RequestURI)
			m.ServeHTTP(w, r)
			return
		}
		log.Println("MULTIPLEX: http. ", r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

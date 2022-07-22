package web

import (
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/quickfeed/quickfeed/internal/rand"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/hooks"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GrpcMultiplexer struct {
	MuxServer *grpcweb.WrappedGrpcServer
}

// GRPCServerWithCredentials starts a new gRPC server with credentials
// generated from TLS certificates.
func GRPCServerWithCredentials(opt grpc.ServerOption, certFile, certKey string) (*grpc.Server, error) {
	// Generate TLS credentials from certificates
	cred, err := credentials.NewServerTLSFromFile(certFile, certKey)
	if err != nil {
		return nil, err
	}
	return grpc.NewServer(grpc.Creds(cred), opt), nil
}

// GRPCServer starts a new server without TLS certificates.
// This server should only be used in combination with an envoy proxy
// that manages the TLS session.
func GRPCServer(opt grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(opt)
}

// MuxHandler routes HTTP and gRPC requests.
func (m *GrpcMultiplexer) MuxHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.MuxServer.IsGrpcWebRequest(r) {
			m.MuxServer.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RegisterRouter registers http endpoints for authentication API and scm provider webhooks.
func (s *QuickFeedService) RegisterRouter(authConfig *oauth2.Config, mux GrpcMultiplexer, public string) *http.ServeMux {
	// Serve static files.
	router := http.NewServeMux()
	assets := http.FileServer(http.Dir(public + "/assets"))
	dist := http.FileServer(http.Dir(public + "/dist"))

	router.Handle("/", mux.MuxHandler(http.StripPrefix("/", assets)))
	router.Handle("/assets/", mux.MuxHandler(http.StripPrefix("/assets/", assets)))
	router.Handle("/static/", mux.MuxHandler(http.StripPrefix("/static/", dist)))

	// Register auth endpoints.
	callbackSecret := rand.String()
	router.HandleFunc("/auth/", auth.OAuth2Login(s.logger, authConfig, callbackSecret))
	// TODO(vera): temporary hack to support teacher scopes, will be removed when OAuth app replaced with GitHub app.
	router.HandleFunc("/auth/teacher/", auth.OAuth2Login(s.logger, authConfig, callbackSecret))
	router.HandleFunc("/auth/callback/", auth.OAuth2Callback(s.logger, s.db, authConfig, s.scms, callbackSecret))
	router.HandleFunc("/logout", auth.OAuth2Logout(s.logger))

	// Register hooks.
	ghHook := hooks.NewGitHubWebHook(s.logger, s.db, s.runner, s.bh.Secret)
	router.HandleFunc("/hook/", ghHook.Handle())

	return router
}

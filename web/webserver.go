package web

import (
	"log"
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/quickfeed/quickfeed/internal/rand"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/hooks"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GrpcMultiplexer struct {
	MuxServer *grpcweb.WrappedGrpcServer
}

// ServerWithCredentials starts a new gRPC server with credentials
// generated from TLS certificates.
func ServerWithCredentials(logger *zap.Logger, certFile, certKey string) (*grpc.Server, error) {
	// Generate TLS credentials from certificates

	cred, err := credentials.NewServerTLSFromFile(certFile, certKey)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer(
		grpc.Creds(cred),
		grpc.ChainUnaryInterceptor(
			auth.UserVerifier(),
			qf.Interceptor(logger),
		),
	)
	return s, nil
}

// MuxHandler routes HTTP and gRPC requests.
func (m *GrpcMultiplexer) MuxHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.MuxServer.IsGrpcWebRequest(r) {
			log.Printf("MUX: got GRPC request: %v", r)
			m.MuxServer.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RegisterRouter registers http endpoints for authentication API and scm provider webhooks.
func (s *QuickFeedService) RegisterRouter(authConfig *oauth2.Config, scms *auth.Scms, mux GrpcMultiplexer, public string) *http.ServeMux {
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
	router.HandleFunc("/auth/callback/", auth.OAuth2Callback(s.logger, s.db, authConfig, scms, callbackSecret))
	router.HandleFunc("/logout", auth.OAuth2Logout(s.logger))

	// Register hooks.
	ghHook := hooks.NewGitHubWebHook(s.logger, s.db, s.runner, s.bh.Secret)
	router.HandleFunc("/hook/", ghHook.Handle())

	return router
}

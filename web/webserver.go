package web

import (
	"fmt"
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/quickfeed/quickfeed/internal/rand"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/hooks"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const QuickFeedServiceName = "qf.QuickFeedService"

type GrpcMultiplexer struct {
	MuxServer *grpcweb.WrappedGrpcServer
}

// GRPCServerWithCredentials starts a new gRPC server with credentials
// generated from TLS certificates.
func GRPCServerWithCredentials(certFile, certKey string, opts ...grpc.ServerOption) (*grpc.Server, error) {
	// Generate TLS credentials from certificates
	cred, err := credentials.NewServerTLSFromFile(certFile, certKey)
	if err != nil {
		return nil, err
	}
	opts = append(opts, grpc.Creds(cred))
	return grpc.NewServer(opts...), nil
}

// GRPCServer starts a new server without TLS certificates.
// This server should only be used in combination with an envoy proxy
// that manages the TLS session.
func GRPCServer(opts ...grpc.ServerOption) *grpc.Server {
	return grpc.NewServer(opts...) // skipcq: GO-S0902
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
func (s *QuickFeedService) RegisterRouter(tm *auth.TokenManager, authConfig *oauth2.Config, mux GrpcMultiplexer, public string) *http.ServeMux {
	// Serve static files.
	router := http.NewServeMux()
	assets := http.FileServer(http.Dir(public + "/assets"))
	dist := http.FileServer(http.Dir(public + "/dist"))

	router.Handle("/", mux.MuxHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, public+"/assets/index.html")
	})))
	router.Handle(auth.Assets, mux.MuxHandler(http.StripPrefix(auth.Assets, assets)))
	router.Handle(auth.Static, mux.MuxHandler(http.StripPrefix(auth.Static, dist)))

	// Register auth endpoints.
	callbackSecret := rand.String()
	router.HandleFunc(auth.Auth, auth.OAuth2Login(s.logger, authConfig, callbackSecret))
	router.HandleFunc(auth.Callback, auth.OAuth2Callback(s.logger, s.db, tm, authConfig, callbackSecret))
	router.HandleFunc(auth.Logout, auth.OAuth2Logout())

	// Register hooks.
	ghHook := hooks.NewGitHubWebHook(s.logger, s.db, s.scmMgr, s.runner, s.bh.Secret)
	router.HandleFunc(auth.Hook, ghHook.Handle())

	return router
}

func VerifyAccessControlMethods(s *grpc.Server) error {
	qfServiceInfo, ok := s.GetServiceInfo()[QuickFeedServiceName]
	if !ok {
		return fmt.Errorf("gRPC server missing %s service", QuickFeedServiceName)
	}
	access := interceptor.GetAccessTable()
	if len(qfServiceInfo.Methods) != len(access) {
		return fmt.Errorf("incorrect number of methods in access control table. Expected: %d, got %d", len(qfServiceInfo.Methods), len(access))
	}
	for _, method := range qfServiceInfo.Methods {
		if _, ok := access[method.Name]; !ok {
			return fmt.Errorf("missing method in access control table: %s", method.Name)
		}
	}
	return nil
}

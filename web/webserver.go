package web

import (
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/rand"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/hooks"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"golang.org/x/oauth2"
)

func (s *QuickFeedService) NewQuickFeedHandler(tm *auth.TokenManager) (string, http.Handler) {
	interceptors := connect.WithInterceptors(
		interceptor.NewMetricsInterceptor(),
		interceptor.NewValidationInterceptor(s.logger),
		interceptor.NewTokenAuthInterceptor(s.logger, tm, s.db),
		interceptor.NewUserInterceptor(s.logger, tm),
		interceptor.NewAccessControlInterceptor(tm),
		interceptor.NewTokenInterceptor(tm),
	)
	return qfconnect.NewQuickFeedServiceHandler(s, interceptors)
}

// RegisterRouter registers http endpoints for authentication API and scm provider webhooks.
func (s *QuickFeedService) RegisterRouter(tm *auth.TokenManager, authConfig *oauth2.Config, public string) *http.ServeMux {
	// Serve static files.
	router := http.NewServeMux()
	assets := http.FileServer(http.Dir(public + "/assets")) // skipcq: GO-S1034
	dist := http.FileServer(http.Dir(public + "/dist"))     // skipcq: GO-S1034

	router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, public+"/assets/index.html")
	}))
	paths, handler := s.NewQuickFeedHandler(tm)
	router.Handle(paths, controller(handler))

	router.Handle(auth.Assets, http.StripPrefix(auth.Assets, assets))
	router.Handle(auth.Static, http.StripPrefix(auth.Static, dist))
	// Register auth endpoints.
	callbackSecret := rand.String()
	router.HandleFunc(auth.Auth, auth.OAuth2Login(s.logger, authConfig, callbackSecret))
	router.HandleFunc(auth.Callback, auth.OAuth2Callback(s.logger, s.db, tm, authConfig, callbackSecret))
	router.HandleFunc(auth.Logout, auth.OAuth2Logout())

	// Register hooks.
	ghHook := hooks.NewGitHubWebHook(s.logger, s.db, s.scmMgr, s.runner, s.bh.Secret, s.streams)
	router.HandleFunc(auth.Hook, ghHook.Handle())

	return router
}

// controller is a wrapper for the QuickFeedService handler that sets a write deadline for the submission stream.
// TODO: Remove this when connect-go finally supports deadlines.
func controller(h http.Handler) http.Handler {
	timeout := 15 * time.Minute
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/qf.QuickFeedService/SubmissionStream" {
			control := http.NewResponseController(w)
			control.SetWriteDeadline(time.Now().Add(timeout))
		}
		h.ServeHTTP(w, r)
	})
}

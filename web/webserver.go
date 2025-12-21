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
)

const (
	// streamTimeout is the timeout for the submission stream.
	streamTimeout = 15 * time.Minute
)

func (s *QuickFeedService) NewQuickFeedHandler() (string, http.Handler) {
	interceptors := connect.WithInterceptors(
		interceptor.NewMetricsInterceptor(),
		interceptor.NewValidationInterceptor(s.logger),
		interceptor.NewTokenAuthInterceptor(s.logger, s.tm, s.db),
		interceptor.NewUserInterceptor(s.logger, s.tm),
		interceptor.NewAccessControlInterceptor(s.tm),
		interceptor.NewTokenInterceptor(s.tm),
	)
	return qfconnect.NewQuickFeedServiceHandler(s, interceptors)
}

// RegisterRouter registers http endpoints for authentication API and scm provider webhooks.
func (s *QuickFeedService) RegisterRouter(webHookSecret, public string) *http.ServeMux {
	// Serve static files.
	router := http.NewServeMux()
	assets := http.FileServer(http.Dir(public + "/assets")) // skipcq: GO-S1034
	dist := http.FileServer(http.Dir(public + "/dist"))     // skipcq: GO-S1034

	router.Handle("/robots.txt", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, public+"/assets/robots.txt")
	}))
	router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, public+"/assets/index.html")
	}))
	paths, handler := s.NewQuickFeedHandler()
	router.Handle(paths, controller(handler, streamTimeout))

	router.Handle(auth.Assets, http.StripPrefix(auth.Assets, assets))
	router.Handle(auth.Static, http.StripPrefix(auth.Static, dist))

	// Register auth endpoints.
	callbackSecret := rand.String()
	authConfig := auth.NewGitHubConfig(s.scmMgr.Config)
	router.HandleFunc(auth.Auth, auth.OAuth2Login(s.logger, authConfig, callbackSecret))
	router.HandleFunc(auth.Callback, auth.OAuth2Callback(s.logger, s.db, s.tm, authConfig, callbackSecret))
	router.HandleFunc(auth.Logout, auth.OAuth2Logout())

	// Register hooks.
	ghHook := hooks.NewGitHubWebHook(s.logger, s.db, s.scmMgr, s.runner, webHookSecret, s.streams, s.tm)
	router.HandleFunc(auth.Hook, ghHook.Handle())

	return router
}

// controller is a wrapper for the QuickFeedService handler that sets a write deadline for the submission stream.
// TODO: Remove this when connect-go finally supports deadlines.
// TODO: https://github.com/connectrpc/connect-go/issues/604
func controller(h http.Handler, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == qfconnect.QuickFeedServiceSubmissionStreamProcedure {
			control := http.NewResponseController(w)
			_ = control.SetWriteDeadline(time.Now().Add(timeout))
		}
		h.ServeHTTP(w, r)
	})
}

package web

import (
	"net/http"

	"github.com/bufbuild/connect-go"
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
	router.Handle(s.NewQuickFeedHandler(tm))
	router.Handle(auth.Assets, http.StripPrefix(auth.Assets, assets))
	router.Handle(auth.Static, http.StripPrefix(auth.Static, dist))

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

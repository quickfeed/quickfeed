package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/autograde/aguis"
	"github.com/autograde/aguis/web/handlers"
	"github.com/go-kit/kit/log"
	h "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/bitbucket"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
)

func main() {
	var (
		httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")
		public   = flag.String("http.public", "public", "directory to server static files from")

		baseURL = flag.String("service.url", "localhost", "service base url")
	)
	flag.Parse()

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	tsLogger := log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(tsLogger, "src", log.DefaultCaller)

	store := sessions.NewCookieStore(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32),
	)
	gothic.Store = store

	// TODO: Only register if env set.
	goth.UseProviders(
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), getCallbackURL(*baseURL, "github")),
		bitbucket.New(os.Getenv("BITBUCKET_KEY"), os.Getenv("BITBUCKET_SECRET"), getCallbackURL(*baseURL, "bitbucket")),
		gitlab.New(os.Getenv("GITLAB_KEY"), os.Getenv("GITLAB_SECRET"), getCallbackURL(*baseURL, "gitlab")),
	)

	sessionStore := aguis.NewSessionStore(store, "authsession")

	db, err := aguis.NewStructOnFileDB(tempFile("agdb.db"), false, logger)

	if err != nil {
		panic(fmt.Sprintf("could not connect to db: %s", err))
	}

	r := mux.NewRouter()

	r.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		sessionStore.Logout(w, r)
	})

	auth := r.PathPrefix("/auth/").Subrouter()
	auth.Handle("/{provider}", handlers.AuthHandler(db, sessionStore))
	auth.Handle("/{provider}/callback", handlers.AuthCallbackHandler(db, sessionStore))

	api := r.PathPrefix("/api/v1/").Subrouter()
	api.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("api call"))
	})

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(*public)))

	srv := &http.Server{
		Handler: h.LoggingHandler(
			loggingHandlerAdapter{
				logger: tsLogger,
				key:    "http",
			},
			handlers.AuthenticatedHandler(r, sessionStore),
		),
		Addr: *httpAddr,
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(fmt.Sprintf("http server error: %s", err))
	}
}

func getCallbackURL(baseURL string, provider string) string {
	return "https://" + baseURL + "/auth/" + provider + "/callback"
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}

func tempFile(name string) string {
	return os.TempDir() + string(filepath.Separator) + name
}

type loggingHandlerAdapter struct {
	logger log.Logger
	key    string
}

func (l loggingHandlerAdapter) Write(p []byte) (int, error) {
	if err := l.logger.Log(l.key, string(p)); err != nil {
		return 0, err
	}
	return len(p), nil
}

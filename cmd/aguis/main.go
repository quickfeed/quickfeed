package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/autograde/aguis"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/handlers"
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
	auth.Handle("/{provider}", authHandler(db, sessionStore))
	auth.Handle("/{provider}/callback", authCallbackHandler(db, sessionStore))

	api := r.PathPrefix("/api/v1/").Subrouter()
	api.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("api call"))
	})

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(*public)))

	srv := &http.Server{
		Handler: handlers.LoggingHandler(
			loggingHandlerAdapter{
				logger: tsLogger,
				key:    "http",
			},
			authenticatedHandler(r, sessionStore),
		),
		Addr: *httpAddr,
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(fmt.Sprintf("http server error: %s", err))
	}
}

// Try to get the user without re-authenticating.
func tryAuthenticate(
	w http.ResponseWriter, r *http.Request,
	db aguis.UserDatabase, s *aguis.Session,
) (*goth.User, error) {
	user, err := gothic.CompleteUserAuth(w, r)

	if err != nil {
		return nil, err
	}

	switch user.Provider {
	case "github":
		if err := loginGithub(db, user.UserID); err != nil {
			return nil, err
		}
		if err := s.Login(w, r); err != nil {
			return nil, err
		}
		return &user, nil
	default:
		return nil, errors.New(user.Provider + " provider not implemented")
	}
}

func authHandler(db aguis.UserDatabase, s *aguis.Session) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := tryAuthenticate(w, r, db, s)
		if err != nil {
			gothic.BeginAuthHandler(w, r)
		}
		serveInfo(w, user)
	})
}

func authCallbackHandler(db aguis.UserDatabase, s *aguis.Session) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := tryAuthenticate(w, r, db, s)
		if err != nil {
			httpError(w, http.StatusInternalServerError, err)
			return
		}
		serveInfo(w, user)
	})
}

func authenticatedHandler(m *mux.Router, s *aguis.Session) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggedIn, err := s.LoggedIn(w, r)

		if err != nil {
			httpError(w, http.StatusInternalServerError, err)
			return
		}

		if strings.HasPrefix(r.RequestURI, "/api") && !loggedIn {
			httpError(w, http.StatusForbidden, nil)
			return
		}
		m.ServeHTTP(w, r)
	})
}

func loginGithub(db aguis.UserDatabase, userID string) error {
	githubID, err := strconv.Atoi(userID)
	if err != nil {
		return err
	}
	_, err = db.GetUserWithGithubID(githubID)
	if err != nil {
		return err
	}
	return nil
}

func serveInfo(w http.ResponseWriter, user *goth.User) {
	t, _ := template.New("").Parse(`
	<p><a href="/logout">logout</a></p>
	<p>Name: {{.Name}}</p>
	<p>NickName: {{.NickName}}</p>
	<p>UserID: {{.UserID}}</p>
	<p>AccessToken: {{.AccessToken}}</p>
	`)

	t.Execute(w, user)
}

func httpError(w http.ResponseWriter, code int, err error) {
	res := http.StatusText(code)
	if err != nil && debug {
		res = fmt.Sprintf("%s: %s", http.StatusText(code), err.Error())
	}
	http.Error(w, res, code)
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

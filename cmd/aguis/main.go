package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/autograde/aguis"
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

	r := mux.NewRouter()
	r.Handle("/", http.FileServer(http.Dir(*public)))

	r.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		sessionStore.Logout(w, r)
	})

	auth := r.PathPrefix("/auth/").Subrouter()
	auth.Handle("/{provider}", authHandler(sessionStore))
	auth.Handle("/{provider}/callback", authCallbackHandler(sessionStore))

	api := r.PathPrefix("/api/v1/").Subrouter()
	api.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("api call"))
	})

	srv := &http.Server{
		Handler: handlers.LoggingHandler(
			os.Stdout,
			authenticatedHandler(r, sessionStore),
		),
		Addr: *httpAddr,
	}

	log.Fatal(srv.ListenAndServe())
}

func authHandler(s *aguis.Session) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get the user without re-authenticating.
		if user, err := gothic.CompleteUserAuth(w, r); err == nil {
			s.Login(w, r, user.AccessToken)
			serveInfo(w, user)
			return
		}

		gothic.BeginAuthHandler(w, r)
	})
}

func authCallbackHandler(s *aguis.Session) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			http.Error(
				w,
				httpError(http.StatusInternalServerError, err),
				http.StatusInternalServerError,
			)
			return
		}
		s.Login(w, r, user.AccessToken)
		serveInfo(w, user)
	})
}

func authenticatedHandler(m *mux.Router, s *aguis.Session) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggedIn, err := s.LoggedIn(w, r)

		if err != nil {
			http.Error(
				w,
				httpError(http.StatusInternalServerError, err),
				http.StatusInternalServerError,
			)
			return
		}

		if strings.HasPrefix(r.RequestURI, "/api") && !loggedIn {
			http.Error(
				w,
				http.StatusText(http.StatusForbidden),
				http.StatusForbidden,
			)
			return
		}
		m.ServeHTTP(w, r)
	})
}

func httpError(code int, err error) string {
	if !debug {
		return http.StatusText(code)
	}
	return fmt.Sprintf("%s: %s", http.StatusText(code), err.Error())
}

func serveInfo(w http.ResponseWriter, user goth.User) {
	t, _ := template.New("").Parse(`
	<p><a href="/logout">logout</a></p>
	<p>Name: {{.Name}}</p>
	<p>NickName: {{.NickName}}</p>
	<p>UserID: {{.UserID}}</p>
	<p>AccessToken: {{.AccessToken}}</p>
	`)

	t.Execute(w, user)
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

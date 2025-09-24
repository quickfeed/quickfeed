package auth

import (
	"time"
)

type contextKey int

const (
	contextNone contextKey = iota
	ContextKeyClaims
)

const (
	Cookie               = "cookie"
	CookieName           = "auth"
	SetCookie            = "Set-Cookie"
	tokenExpirationTime  = 15 * time.Minute
	cookieExpirationTime = 24 * time.Hour * 14 // 2 weeks
	alg                  = "HS256"

	// nextCookieName is the name of the cookie that stores the next URL to redirect to after login.
	nextCookieName = "qf_next"

	githubUserAPI = "https://api.github.com/user"

	// Routes
	Auth     = "/auth/"
	Teacher  = "/auth/teacher/"
	Callback = "/auth/callback/"
	Logout   = "/logout"
	Hook     = "/hook/"
	Assets   = "/assets/"
	Static   = "/static/"
)

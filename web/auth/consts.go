package auth

import "time"

const (
	Cookie               = "cookie"
	CookieName           = "auth"
	UserKey              = "user"
	SetCookie            = "Set-Cookie"
	tokenExpirationTime  = 15 * time.Minute
	cookieExpirationTime = 12 * time.Hour
	alg                  = "HS256"

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

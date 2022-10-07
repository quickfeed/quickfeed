package auth

import (
	"context"
	"time"

	"github.com/quickfeed/quickfeed/qf"
)

type contextKey int

const (
	contextNone contextKey = iota
	ContextKeyUserID
	ContextKeyClaims
)

const (
	Cookie               = "cookie"
	CookieName           = "auth"
	SetCookie            = "Set-Cookie"
	tokenExpirationTime  = 15 * time.Minute
	cookieExpirationTime = 24 * time.Hour * 14 // 2 weeks
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

// WithUserContext returns the context augmented with the given user's ID.
// This aims to mimic the claims.Context() method.
func WithUserContext(ctx context.Context, user *qf.User) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, user.GetID())
}

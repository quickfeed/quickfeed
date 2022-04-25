package config

import "time"

const (
	// Environmental variables
	WebhookEnv    = "WEBHOOK_SECRET"
	JWTKeyEnv     = "JWT_KEY" // TODO(vera): where to store? Or reuse another secret?
	JWTCookieName = "auth"
	// Endpoints
	Install  = "https://github.com/apps/appth-gh" // TODO(vera): change to the real URL (or better yet read from a config file)
	Login    = "/auth/github"
	Logout   = "/logout"
	Callback = "/auth/github/callback"
	Webhook  = "/hook/github/events"

	// Paths //TODO(vera): read from env
	pemPath   = "cert/server.crt"
	keyPath   = "cert/server.key"
	indexFile = "index.html"

	TokenExpirationTime = time.Hour * 244
)

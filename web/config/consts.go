package config

import "time"

const (
	// Environmental variables
	WebhookEnv  = "WEBHOOK_SECRET"
	TokenKeyEnv = "JWT_KEY" // TODO(vera): where to store? Or reuse another secret?
	CertEnv     = "CERT"
	CertKeyEnv  = "CERT_KEY"
	// Endpoints
	Install = "https://github.com/apps/appth-gh" // TODO(vera): change to the real URL (or better yet read from a config file)
	Login   = "/auth/github/"
	// Logout   = "/logout" TODO(vera): do we need an explicit logout somewhere?
	Callback = "/auth/github/callback/"
	Webhook  = "/hook/github/events"

	indexFile      = "index.html"
	AuthCookieName = "auth"

	TokenExpirationTime = time.Hour * 244
)

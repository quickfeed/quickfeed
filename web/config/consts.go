package config

import "time"

const (
	// Environmental variables
	WebhookEnv  = "WEBHOOK_SECRET"
	TokenKeyEnv = "JWT_KEY"
	CertEnv     = "CERT"
	CertKeyEnv  = "CERT_KEY"
	// Endpoints
	Install  = "https://github.com/apps/appth-gh" // TODO(vera): change to the real URL (or better yet read from a config file)
	Login    = "/auth/github/"
	Callback = "/auth/github/callback/"
	Webhook  = "/hook/github/events"

	TokenExpirationTime = time.Hour * 244
)

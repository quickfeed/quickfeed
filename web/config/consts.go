package config

const (
	// Environmental variables
	WebhookEnv = "WEBHOOK_SECRET"
	JWTKeyEnv  = "JWT_KEY" // TODO: where to store? Or reuse another secret?

	// Endpoints
	GitHubUser = "https://api.github.com/user"
	Install    = "https://github.com/apps/appth-gh" // TODO: change to the real URL (or better yet read from a config file)
	Login      = "/login"
	Logout     = "/logout"
	Callback   = "/callback"
	Webhook    = "/hook/github/events"

	// Paths //TODO: read from env
	pemPath   = "cert/server.crt"
	keyPath   = "cert/server.key"
	indexFile = "index.html"
)

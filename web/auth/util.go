package auth

import "github.com/quickfeed/quickfeed/internal/env"

// GetCallbackURL returns the callback URL for a given base URL and a provider.
func GetCallbackURL() string {
	return getURL(env.DomainWithPort(), Callback)
}

// GetEventsURL returns the event URL for a given base URL and a provider.
func GetEventsURL() string {
	return getURL(env.Domain(), Hook)
}

// GetBaseURL returns the base URL.
func GetBaseURL() string {
	return getURL(env.Domain(), "")
}

// getURL constructs an URL endpoint for the given route.
func getURL(baseURL, route string) string {
	return "https://" + baseURL + route
}

// externalUser is used to decode user authentication JSON sent as response by OAuth providers.
type externalUser struct {
	ID        uint64 `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

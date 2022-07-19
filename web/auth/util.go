package auth

// GetCallbackURL returns the callback URL for a given base URL and a provider.
func GetCallbackURL(baseURL, provider string) string {
	return GetProviderURL(baseURL, "auth/callback", provider)
}

// GetEventsURL returns the event URL for a given base URL and a provider.
func GetEventsURL(baseURL, provider string) string {
	return GetProviderURL(baseURL, "hook", provider)
}

// GetProviderURL returns a URL endpoint given a base URL and a provider.
func GetProviderURL(baseURL, route, provider string) string {
	return "https://" + baseURL + "/" + route + "/" + provider // + "/" + endpoint
}

// externalUser is used to decode the user authentication response from OAuth providers.
type externalUser struct {
	ID        uint64 `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

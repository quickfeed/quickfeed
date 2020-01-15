package auth

// GetCallbackURL returns the callback URL for a given base URL and a provider.
func GetCallbackURL(baseURL, provider string) string {
	return GetProviderURL(baseURL, "auth", provider, "callback")
}

// GetEventsURL returns the event URL for a given base URL and a provider.
func GetEventsURL(baseURL, provider string) string {
	return GetProviderURL(baseURL, "hook", provider, "events")
}

// GetProviderURL returns a URL endpoint given a base URL and a provider.
func GetProviderURL(baseURL, route, provider, endpoint string) string {
	return "https://" + baseURL + "/" + route + "/" + provider + "/" + endpoint
}

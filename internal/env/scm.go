package env

import "os"

const (
	defaultProvider = "github"
)

var provider string

func init() {
	provider = os.Getenv("QUICKFEED_SCM_PROVIDER")
	if provider == "" {
		provider = defaultProvider
	}
}

// ScmProvider returns the current SCM provider supported by this backend.
func ScmProvider() string {
	return provider
}

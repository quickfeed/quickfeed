package ci

// Language constants for supported languages.
// These are used in run scripts via the #language/ directive.
const (
	languageGo     = "go"
	languageDotNet = "dotnet"
)

// Language-specific container cache paths.
const (
	GoModCache        = "/quickfeed-go-mod-cache"
	GoCache           = "/quickfeed-go-cache"
	GolangciLintCache = "/quickfeed-golangci-lint-cache"
	NuGetCache        = "/quickfeed-nuget-cache"
)

// languageConfig defines language-specific cache mounts and environment variables.
type languageConfig struct {
	// cacheDirs maps container target paths to host cache directory resolver functions.
	cacheDirs map[string]func() (string, error)
	// envVars lists additional environment variables to set in the container.
	envVars []string
}

// languages maps language identifiers to their cache configurations.
// To add support for a new language, add an entry here with the relevant
// cache directories and environment variables.
var languages = map[string]languageConfig{
	languageGo: {
		cacheDirs: map[string]func() (string, error){
			GoModCache:        moduleCachePath,
			GoCache:           goCachePath,
			GolangciLintCache: golangciLintCachePath,
		},
		envVars: []string{
			"GOMODCACHE=" + GoModCache,
			"GOCACHE=" + GoCache,
			"GOLANGCI_LINT_CACHE=" + GolangciLintCache,
		},
	},
	languageDotNet: {
		cacheDirs: map[string]func() (string, error){
			NuGetCache: nugetCachePath,
		},
		envVars: []string{
			"NUGET_PACKAGES=" + NuGetCache,
		},
	},
}

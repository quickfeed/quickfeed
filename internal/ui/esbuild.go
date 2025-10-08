package ui

import (
	"fmt"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/quickfeed/quickfeed/internal/env"
	"path/filepath"
)

var (
	public = func(s string) string {
		return filepath.Join(env.PublicDir(), s)
	}
	distDir = public("dist")
)

// buildOptions defines the build options for esbuild
// The api has write access and writes the output to public/dist
var buildOptions = api.BuildOptions{
	Outdir:            distDir,
	EntryPoints:       entryPoints,
	Bundle:            true,
	Write:             true,
	TreeShaking:       api.TreeShakingTrue, // Remove unused code
	EntryNames:        "[name]-[hash]",     // Use hash to ensure the user always gets the latest version
	Splitting:         true,
	Format:            api.FormatESModule,
	MinifyWhitespace:  true,
	MinifyIdentifiers: true,
	MinifySyntax:      true,
	LogLevel:          api.LogLevelError,
	LogOverride: map[string]api.LogLevel{
		"unsupported-dynamic-import": api.LogLevelSilent,
	},
	Sourcemap: api.SourceMapLinked,
	Loader: map[string]api.Loader{
		".scss": api.LoaderCSS, // Treat SCSS files as CSS
	},
	Plugins: plugins,
}

var entryPoints = []string{
	public("src/index.tsx"),
	public("src/App.tsx"),

	// pages
	public("src/pages/TeacherPage.tsx"),

	// components
	public("src/components/manual-grading/Comment.tsx"),
	public("src/components/Card.tsx"),

	// overmind
	public("src/overmind/index.ts"),
	public("src/overmind/namespaces/global/effects.ts"),
	public("src/overmind/state.ts"),
	public("src/overmind/namespaces/global/internalActions.ts"),
}

// getOptions returns the build options for esbuild
// used to perform dynamic updates depending on the dev flag and outputDir
func getOptions(outputDir string, dev bool) api.BuildOptions {
	// its important to call env.GetAppURL after the env variable is loaded
	buildOptions.Define = map[string]string{
		"process.env.QUICKFEED_APP_URL": fmt.Sprintf("%q", env.GetAppURL()),
	}
	if dev {
		// Esbuild defaults to production when minifying files.
		// We must explicitly set it to "development" for dev builds.
		buildOptions.Define["process.env.NODE_ENV"] = `"development"`
		buildOptions.LogLevel = api.LogLevelDebug
	}
	// enabling custom outputDir allow for testing without overwriting current build
	if outputDir != "" {
		buildOptions.Outdir = outputDir
	}
	return buildOptions
}

// Build builds the UI with esbuild and tailwind. If outputDir is an empty string, it defaults to public/dist.
// Test cases should pass a non-empty outputDir to avoid overwriting the current build.
func Build(outputDir string, dev bool) error {
	result := api.Build(getOptions(outputDir, dev))
	if len(result.Errors) > 0 {
		return fmt.Errorf("failed to build UI: %v", result.Errors)
	}
	return nil
}

// Watch starts a watch process for both tailwind and esbuild, rebuilding on changes
func Watch() error {
	ctx, err := api.Context(getOptions("", true))
	if err != nil {
		return fmt.Errorf("failed to create build context: %w", err)
	}
	return ctx.Watch(api.WatchOptions{})
}

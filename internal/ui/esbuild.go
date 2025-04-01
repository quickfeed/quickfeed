package ui

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/quickfeed/quickfeed/internal/env"
)

/*

	Esbuild currently struggles with splitting the compiled files into reasonable sized chunks.
	Therefore will webpack be used until esbuild can handle this.
	There are improvements which can be made to better split the files into chunks.
	But the splitting functionality is still experimental in esbuild, and we therefore wait to use it in production.

*/

// buildOptions defines the build options for esbuild
// The api has write access and writes the output to public/dist
var buildOptions = api.BuildOptions{
	Outdir:      fmt.Sprintf("%s/dist", env.PublicDir()),
	Bundle:      true,
	Write:       true,
	TreeShaking: api.TreeShakingTrue,
	EntryNames:  "[name]-[hash]",
	Splitting:   true,
	Metafile:    true,
	Format:      api.FormatESModule,
	Loader: map[string]api.Loader{
		".scss": api.LoaderCSS, // Treat SCSS files as CSS
	},
}

// getOptions returns the build options for esbuild
func getOptions(outputDir *string) api.BuildOptions {
	// Dynamically add entry points to the build options
	// Adding more can reduce the size of the output files, but will increase the file count
	entryPoints := []string{
		"src/index.tsx",
		"src/App.tsx",
		"src/overmind/state.tsx",
		"src/overmind/actions.tsx",
		"src/overmind/effects.tsx",
		"src/overmind/index.tsx",
		"src/components/Results.tsx",
	}
	for _, entry := range entryPoints {
		buildOptions.EntryPoints = append(buildOptions.EntryPoints, fmt.Sprintf("%s/%s", env.PublicDir(), entry))
	}
	/*
		Production options

		These can be added for development, but is not recommended.

		buildOptions.LogLevel = api.LogLevelError
		buildOptions.MinifyWhitespace = true
		buildOptions.MinifyIdentifiers = true
		buildOptions.MinifySyntax = true
	*/

	// Development options
	buildOptions.LogLevel = api.LogLevelDebug
	buildOptions.Sourcemap = api.SourceMapInline

	// This is done to enable testing without overwriting current build
	// Important to not override webpack build
	if outputDir != nil {
		buildOptions.Outdir = *outputDir
	}
	return buildOptions
}

// resetDistFolder removes the dist folder and creates a new one
func resetDistFolder() error {
	path := fmt.Sprintf("%s/dist", env.PublicDir())
	if _, err := os.Stat(path); err == nil {
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove dist directory: %v", err)
		}
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("failed to create dist directory: %v", err)
	}
	return nil
}

// Build builds the UI with esbuild
// The entry point is src/index.tsx and the output is public/dist
// Scss files are treated as css
func Build(outputDir *string) error {
	if err := resetDistFolder(); err != nil {
		return fmt.Errorf("failed to reset dist folder: %v", err)
	}
	result := api.Build(getOptions(outputDir))
	if len(result.Errors) > 0 {
		return fmt.Errorf("failed to build UI: %v", parseMessages(result.Errors))
	}
	if err := createHtml(result.OutputFiles); err != nil {
		return fmt.Errorf("failed to create index.html: %v", err)
	}
	return nil
}

// createHtml creates the index.html file for the UI
// Injects file links into the index template
func createHtml(outputFiles []api.OutputFile) error {
	file, err := os.Create(fmt.Sprintf("%s/assets/index.html", env.PublicDir()))
	if err != nil {
		return fmt.Errorf("failed to read index.html: %v", err)
	}
	funcMap := template.FuncMap{
		"ext":  filepath.Ext,
		"base": filepath.Base,
	}
	tmpl, err := os.ReadFile(fmt.Sprintf("%s/index.esbuild.tmpl.html", env.PublicDir()))
	if err != nil {
		return fmt.Errorf("failed to read index.tmpl.html: %v", err)
	}
	t, err := template.New("index.html").Funcs(funcMap).Parse(string(tmpl))
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}
	err = t.Execute(file, outputFiles)
	if err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}
	return nil
}

// Watch starts a watch process for the UI, rebuilding on changes
// The log level is set to info, so only warnings and errors are shown
func Watch(ch chan<- error) {
	ctx, err := api.Context(getOptions(nil))
	defer ctx.Cancel()
	if err != nil {
		ch <- fmt.Errorf("failed to create build context: %v", parseMessages(err.Errors))
		return
	}
	if err := ctx.Watch(api.WatchOptions{}); err != nil {
		ch <- fmt.Errorf("failed to start watching: %v", err)
		return
	}
	ch <- nil
}

// parseMessages converts esbuild messages to a single error
func parseMessages(messages []api.Message) error {
	var errs []error
	for _, message := range messages {
		errs = append(errs, fmt.Errorf("error: %s", message.Text))
	}
	return errors.Join(errs...)
}

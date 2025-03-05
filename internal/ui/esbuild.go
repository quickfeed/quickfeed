package ui

import (
	"errors"
	"fmt"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/quickfeed/quickfeed/internal/env"
)

// buildOptions defines the build options for esbuild
// The entry point is src/index.tsx
// The api has write access and writes the output to public/dist
var buildOptions = api.BuildOptions{
	EntryPoints: []string{fmt.Sprintf("%s/src/index.tsx", env.PublicDir())},
	Outdir:      fmt.Sprintf("%s/dist", env.PublicDir()),
	Bundle:      true,
	Write:       true,
	Loader: map[string]api.Loader{
		".scss": api.LoaderCSS, // Treat SCSS files as CSS
	},
}

// getOptions updates the build options based on the dev flag
// Dev mode uses inline source maps, and has a debug log level
// Production mode minifies the output to boost performance, and logs only errors
func getOptions(dev bool, outputDir *string) api.BuildOptions {
	opts := buildOptions
	if dev {
		opts.Sourcemap = api.SourceMapInline
		opts.LogLevel = api.LogLevelDebug
	} else {
		opts.LogLevel = api.LogLevelError
		opts.MinifyWhitespace = true
		opts.MinifyIdentifiers = true
		opts.MinifySyntax = true
	}
	// This is done to enable testing
	if outputDir != nil {
		opts.Outdir = *outputDir
	}
	return opts
}

// Build builds the UI with esbuild
// The entry point is src/index.tsx and the output is public/dist
// Scss files are treated as css
func Build(dev bool, outputDir *string) error {
	result := api.Build(getOptions(dev, outputDir))
	if len(result.Errors) > 0 {
		return fmt.Errorf("failed to build UI: %v", parseMessages(result.Errors))
	}
	return nil
}

// Watch starts a watch process for the UI, rebuilding on changes
// The log level is set to info, so only warnings and errors are shown
func Watch(ch chan<- error, dev bool, outputDir *string) {
	ctx, err := api.Context(getOptions(dev, nil))
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
		errs = append(errs, fmt.Errorf("error: %s, in file: %s", message.Text, message.Location.File))
	}
	return errors.Join(errs...)
}

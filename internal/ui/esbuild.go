package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	Format:      api.FormatESModule,
	Splitting:   true,
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
	// This is done to enable testing without overwriting current build
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
	if err := createHtml(); err != nil {
		return fmt.Errorf("failed to create index.html: %v", err)
	}
	return nil
}

// createIndexTemplate creates the index template for the UI
// It injects script references into the index template
// The index template is located in public/index.tmpl.html
// The script references are located in public/dist
// The index template is written to public/assets/index.html
func createHtml() error {
	indexTmpl := fmt.Sprintf("%s/index.tmpl.html", env.PublicDir())
	f, err := os.ReadFile(indexTmpl)
	if err != nil {
		return fmt.Errorf("failed to stat index template: %v", err)
	}
	lines := removeComments(strings.Split(string(f), "\n"))

	indexToInject := findClosingHeadTag(lines)
	if indexToInject == -1 {
		return errors.New("failed to find </head> in index template")
	}

	links, err := getLinks()
	if err != nil {
		return fmt.Errorf("failed to get script refs: %v", err)
	}

	// Inject script references into index template and write to public/assets/index.html
	lines = append(lines[:indexToInject], append(links, lines[indexToInject:]...)...)
	if err := os.WriteFile(fmt.Sprintf("%s/assets/index.html", env.PublicDir()), []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write index.html: %v", err)
	}
	return nil
}

// findClosingHeadTag attempts to find the closing head tag in the index template
// Returns index of the line before the closing head tag
// The closing head tag is defined as </head>
func findClosingHeadTag(lines []string) int {
	for i, line := range lines {
		if strings.Contains(line, "</head>") {
			return i
		}
	}
	return -1
}

// getLinks gets the script and css references from the dist directory
// Support extensions are .css and .js
func getLinks() ([]string, error) {
	content, err := os.ReadDir(fmt.Sprintf("%s/dist", env.PublicDir()))
	if err != nil {
		return nil, fmt.Errorf("failed to read dist directory: %v", err)
	}
	var links []string
	for _, file := range content {
		if file.IsDir() {
			return nil, errors.New("unexpected directory in dist")
		}
		name := file.Name()
		switch filepath.Ext(name) {
		case ".css":
			links = append(links, fmt.Sprintf("\t<link rel=\"stylesheet\" href=\"/static/%s\">", name))
		case ".js":
			links = append(links, fmt.Sprintf("\t<script type=\"module\" src=\"/static/%s\" defer></script>", name))
		}
	}
	return links, nil
}

// removeComments removes comments from the index template
// Comments are defined as <!-- ... -->
// Appends lines when the comment variable is false with is updated by start: <!-- and end: -->
func removeComments(lines []string) []string {
	var cleaned []string
	var comment bool
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "<!--") {
			comment = true
		}
		if strings.HasSuffix(strings.TrimSpace(line), "-->") {
			comment = false
			continue
		}
		if !comment {
			cleaned = append(cleaned, line)
		}
	}
	return cleaned
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

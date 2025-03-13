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
	if dev {
		buildOptions.Sourcemap = api.SourceMapInline
		buildOptions.LogLevel = api.LogLevelDebug
	} else {
		buildOptions.LogLevel = api.LogLevelError
		buildOptions.MinifyWhitespace = true
		buildOptions.MinifyIdentifiers = true
		buildOptions.MinifySyntax = true
	}
	// This is done to enable testing without overwriting current build
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
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create dist directory: %v", err)
	}
	return nil
}

// Build builds the UI with esbuild
// The entry point is src/index.tsx and the output is public/dist
// Scss files are treated as css
func Build(dev bool, outputDir *string) error {
	if err := resetDistFolder(); err != nil {
		return fmt.Errorf("failed to reset dist folder: %v", err)
	}
	result := api.Build(getOptions(dev, outputDir))
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
	t, err := template.New("index.html").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}
	err = t.Execute(file, outputFiles)
	if err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}
	return nil
}

var tmpl = `<!doctype html>
<html lang="en">

<head>
    <meta charset="utf-8" />
    <link rel="icon" type="image/png" sizes="16x16" href="/assets/img/favicon-16x16.png">
    <link rel="icon" type="image/png" sizes="32x32" href="/assets/img/favicon-32x32.png">
    <link rel="apple-touch-icon" sizes="180x180" href="/assets/img/apple-touch-icon.png">
    <link rel="manifest" href="/assets/site.webmanifest">
    <title>QuickFeed</title>
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css"
        integrity="sha384-JcKb8q3iqJ61gNV9KGb8thSsNjpSL0n8PARn9HuZOnIxN0hoP+VmmDGMN5t9UJ0Z" crossorigin="anonymous">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.min.css">
{{- range $file := . }}
	{{- $ext := (ext $file.Path) -}}
	{{- $name := (base $file.Path) -}}
	{{ with (eq $ext ".css") }}<link rel="stylesheet" href="/static/{{ $name }}">{{ end }}
	{{ with (eq $ext ".js") }}<script type="module" src="/static/{{ $name }}" defer></script>{{ end }}
{{- end }}
</head>

<body>
    <noscript>You need to enable JavaScript to run this app.</noscript>
    <div id="root"></div>
</body>

</html>
`

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
		errs = append(errs, fmt.Errorf("error: %s", message.Text))
	}
	return errors.Join(errs...)
}

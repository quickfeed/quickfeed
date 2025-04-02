package ui

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/quickfeed/quickfeed/internal/env"
)

var public = func(s string) string {
	return fmt.Sprintf("%s/%s", env.PublicDir(), s)
}

// buildOptions defines the build options for esbuild
// The api has write access and writes the output to public/dist
var buildOptions = api.BuildOptions{
	Outdir: fmt.Sprintf("%s/dist", env.PublicDir()),
	EntryPoints: []string{
		public("src/index.tsx"),
		public("src/App.tsx"),
		public("src/overmind/state.tsx"),
		public("src/overmind/actions.tsx"),
		public("src/overmind/effects.tsx"),
		public("src/overmind/index.tsx"),
		public("src/components/Results.tsx"),
		// Adding more can reduce the size of the output files, but will increase the file count
	},
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
	Sourcemap:         api.SourceMapLinked,
	Loader: map[string]api.Loader{
		".scss": api.LoaderCSS, // Treat SCSS files as CSS
	},
	Plugins: []api.Plugin{
		{
			Name: "Reset Plugin",
			Setup: func(setup api.PluginBuild) {
				setup.OnStart(func() (api.OnStartResult, error) {
					err := resetDistFolder()
					if err != nil {
						log.Printf("failed to reset dist folder: %v", err)
					}
					return api.OnStartResult{}, err
				})
			},
		},
		{
			Name: "HTML Plugin",
			Setup: func(setup api.PluginBuild) {
				setup.OnEnd(func(result *api.BuildResult) (api.OnEndResult, error) {
					err := createHtml(result.OutputFiles)
					if err != nil {
						log.Printf("failed to create index.html: %v", err)
					}
					return api.OnEndResult{Errors: result.Errors, Warnings: result.Warnings}, err
				})
			},
		},
	},
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

// createHtml creates the index.html file from the index.tmpl.html template
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
	tmpl, err := os.ReadFile(fmt.Sprintf("%s/index.tmpl.html", env.PublicDir()))
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

// getOptions returns the build options for esbuild
// used to perform dynamic updates depending on the dev flag and outputDir
func getOptions(outputDir *string, dev bool) api.BuildOptions {
	if dev {
		buildOptions.Define = map[string]string{
			"process.env.NODE_ENV": "\"development\"", // Required to define development mode when minifying files, or it will default to production
		}
		buildOptions.LogLevel = api.LogLevelDebug
	}
	// enabling custom outputDir allow for testing without overwriting current build
	if outputDir != nil {
		buildOptions.Outdir = *outputDir
	}
	return buildOptions
}

// Build builds the UI with esbuild and outputs to the public/dist folder
func Build(outputDir *string, dev bool) error {
	result := api.Build(getOptions(outputDir, dev))
	if len(result.Errors) > 0 {
		return fmt.Errorf("failed to build UI: %v", parseMessages(result.Errors))
	}
	return nil
}

// Watch starts a watch process for the frontend, rebuilding on changes
func Watch(ch chan<- error) {
	ctx, err := api.Context(getOptions(nil, true))
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

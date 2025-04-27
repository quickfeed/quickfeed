package ui

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/quickfeed/quickfeed/internal/env"
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
	Outdir: distDir,
	EntryPoints: []string{
		public("src/index.tsx"),
		public("src/App.tsx"),

		// pages
		public("src/pages/TeacherPage.tsx"),

		// components
		public("src/components/manual-grading/Comment.tsx"),
		public("src/components/Card.tsx"),

		// overmind
		public("src/overmind/index.tsx"),
		public("src/overmind/effects.tsx"),
		public("src/overmind/state.tsx"),
		public("src/overmind/internalActions.tsx"),
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
					if err := resetDistFolder(); err != nil {
						return api.OnStartResult{
							Warnings: []api.Message{
								{
									PluginName: "Reset",
									Text:       "Failed to clear the dist folder",
									Notes: []api.Note{
										{Text: fmt.Sprintf("The dist directory may now contain multiple builds\nLocation: %s", distDir)},
										{Text: fmt.Sprintf("Error: %v", err)},
									},
								},
							},
						}, nil
					}
					return api.OnStartResult{}, nil
				})
			},
		},
		{
			Name: "HTML Plugin",
			Setup: func(setup api.PluginBuild) {
				setup.OnEnd(func(result *api.BuildResult) (api.OnEndResult, error) {
					if err := createHtml(result.OutputFiles); err != nil {
						return api.OnEndResult{
							Errors: []api.Message{
								{
									PluginName: "HTML",
									Text:       "Failed to create index.html",
									Notes:      []api.Note{{Text: err.Error()}},
								},
							},
						}, nil
					}
					return api.OnEndResult{}, nil
				})
			},
		},
	},
}

// resetDistFolder removes the dist folder and creates a new one
func resetDistFolder() error {
	entries, err := os.ReadDir(distDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		err = os.Remove(filepath.Join(distDir, entry.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}

// createHtml creates the index.html file from the index.tmpl.html template
// Injects file links into the index template
func createHtml(outputFiles []api.OutputFile) error {
	file, err := os.Create(public("assets/index.html"))
	if err != nil {
		return err
	}
	funcMap := template.FuncMap{
		"ext":  filepath.Ext,
		"base": filepath.Base,
	}
	tmplName := "index.tmpl.html"
	tmpl, err := os.ReadFile(public(tmplName))
	if err != nil {
		return err
	}
	t, err := template.New(tmplName).Funcs(funcMap).Parse(string(tmpl))
	if err != nil {
		return err
	}
	return t.Execute(file, outputFiles)
}

// getOptions returns the build options for esbuild
// used to perform dynamic updates depending on the dev flag and outputDir
func getOptions(outputDir string, dev bool) api.BuildOptions {
	// its important to call env.GetAppURL after the env variable is loaded
	buildOptions.Define = map[string]string{
		"process.env.QUICKFEED_APP_URL": fmt.Sprintf(`"%s"`, env.GetAppURL()),
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

// Build builds the UI with esbuild and outputs to the public/dist folder
func Build(outputDir string, dev bool) error {
	result := api.Build(getOptions(outputDir, dev))
	if len(result.Errors) > 0 {
		return fmt.Errorf("failed to build UI: %v", result.Errors)
	}
	return nil
}

// Watch starts a watch process for the frontend, rebuilding on changes
func Watch() error {
	ctx, err := api.Context(getOptions("", true))
	if err != nil {
		return fmt.Errorf("failed to create build context: %w", err)
	}
	return ctx.Watch(api.WatchOptions{})
}

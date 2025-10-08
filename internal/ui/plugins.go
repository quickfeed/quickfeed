package ui

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/quickfeed/quickfeed/internal/env"
)

var plugins = []api.Plugin{
	{
		Name: "Reset dist folder",
		Setup: func(setup api.PluginBuild) {
			setup.OnStart(func() (api.OnStartResult, error) {
				if err := resetDistFolder(); err != nil {
					return api.OnStartResult{
						Warnings: createMessage("Reset dist folder", "Failed to clear the dist folder", err),
					}, nil
				}
				return api.OnStartResult{}, nil
			})
		},
	},
	{
		// important to run tailwind after clearing the dist folder
		Name: "Tailwind",
		Setup: func(setup api.PluginBuild) {
			setup.OnStart(func() (api.OnStartResult, error) {
				cmd := exec.Command("npm", "run", "tailwind")
				cmd.Dir = env.PublicDir()
				if err := cmd.Run(); err != nil {
					return api.OnStartResult{
						Warnings: createMessage("Tailwind", "Failed to generate Tailwind CSS", err),
					}, nil
				}
				return api.OnStartResult{}, nil
			})
		},
	},
	{
		Name: "HTML",
		Setup: func(setup api.PluginBuild) {
			setup.OnEnd(func(result *api.BuildResult) (api.OnEndResult, error) {
				if err := createHtml(result.OutputFiles); err != nil {
					return api.OnEndResult{
						Errors: createMessage("HTML", "Failed to create index.html", err),
					}, nil
				}
				return api.OnEndResult{}, nil
			})
		},
	},
}

func createMessage(pluginName, text string, err error) []api.Message {
	msg := api.Message{
		PluginName: pluginName,
		Text:       text,
		Notes: []api.Note{
			{Text: fmt.Sprintf("Error: %v", err)},
		},
	}
	return []api.Message{msg}
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

package ui

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
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
		Name: "Check Node Modules",
		Setup: func(api.PluginBuild) {
			if _, err := os.Stat(filepath.Join(env.PublicDir(), "node_modules")); os.IsNotExist(err) {
				fmt.Println("node_modules not found, installing...")
				cmd := exec.Command("npm", "ci")
				cmd.Dir = env.PublicDir()
				if err := cmd.Run(); err != nil {
					fmt.Printf("failed to install node modules: %v\n", err)
				}
			}
		},
	},
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

// htmlData holds the values injected into the index.tmpl.html template.
type htmlData struct {
	// TailwindHash versions the Tailwind stylesheet's URL, since esbuild does not hash it.
	// Without it, browsers may pair freshly deployed JS with a cached stylesheet
	// that lacks the Tailwind classes the new JS uses.
	TailwindHash string
	OutputFiles  []api.OutputFile
}

// createHtml creates the index.html file from the index.tmpl.html template
// Injects file links into the index template
func createHtml(outputFiles []api.OutputFile) error {
	// The Tailwind step only warns when it fails; link the new bundles anyway
	// rather than leave index.html pointing at the ones the rebuild removed.
	tailwindHash, _ := fileHash(filepath.Join(distDir, "tailwind.css"))
	html, err := renderHtml(public("index.tmpl.html"), htmlData{
		TailwindHash: tailwindHash,
		OutputFiles:  outputFiles,
	})
	if err != nil {
		return err
	}
	return os.WriteFile(public("assets/index.html"), html, 0o644)
}

// renderHtml executes the template at tmplPath with data.
func renderHtml(tmplPath string, data htmlData) ([]byte, error) {
	tmpl, err := os.ReadFile(tmplPath)
	if err != nil {
		return nil, err
	}
	funcMap := template.FuncMap{
		"ext":  filepath.Ext,
		"base": filepath.Base,
	}
	t, err := template.New(filepath.Base(tmplPath)).Funcs(funcMap).Parse(string(tmpl))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// fileHash returns a short hash of the file's content.
func fileHash(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:4]), nil
}

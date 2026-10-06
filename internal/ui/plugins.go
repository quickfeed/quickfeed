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
		// esbuild runs OnStart callbacks concurrently, so the dist folder
		// must be cleared and Tailwind generated within a single callback.
		Name: "Prepare dist folder",
		Setup: func(setup api.PluginBuild) {
			setup.OnStart(func() (api.OnStartResult, error) {
				var warnings []api.Message
				if err := resetDistFolder(); err != nil {
					warnings = append(warnings, createMessage("Reset dist folder", "Failed to clear the dist folder", err)...)
				}
				if err := generateTailwind(); err != nil {
					warnings = append(warnings, createMessage("Tailwind", "Failed to generate Tailwind CSS", err)...)
				}
				return api.OnStartResult{Warnings: warnings}, nil
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

// generateTailwind runs the Tailwind CLI, which writes dist/tailwind.css.
func generateTailwind() error {
	cmd := exec.Command("npm", "run", "tailwind")
	cmd.Dir = env.PublicDir()
	return cmd.Run()
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
	// It is empty if the stylesheet could not be read, in which case the URL is left unversioned.
	TailwindHash string
	OutputFiles  []api.OutputFile
}

// createHtml creates the index.html file from the index.tmpl.html template
// Injects file links into the index template
func createHtml(outputFiles []api.OutputFile) error {
	return writeHtml(public("index.tmpl.html"), public("assets/index.html"), filepath.Join(distDir, "tailwind.css"), outputFiles)
}

// writeHtml renders the template at tmplPath to htmlPath,
// versioning the Tailwind stylesheet at tailwindPath by its content.
func writeHtml(tmplPath, htmlPath, tailwindPath string, outputFiles []api.OutputFile) error {
	// The Tailwind step only warns when it fails; link the new bundles anyway
	// rather than leave index.html pointing at the ones the rebuild removed.
	tailwindHash, err := fileHash(tailwindPath)
	if err != nil {
		fmt.Printf("Tailwind stylesheet will not be versioned: %v\n", err)
	}
	html, err := renderHtml(tmplPath, htmlData{
		TailwindHash: tailwindHash,
		OutputFiles:  outputFiles,
	})
	if err != nil {
		return err
	}
	return os.WriteFile(htmlPath, html, 0o644)
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

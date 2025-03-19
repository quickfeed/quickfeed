package main

import (
	"embed"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func addLintCheckers(args []string) {
	fs := flag.NewFlagSet(addLintCheckersCmd, flag.ExitOnError)
	var labs string
	fs.StringVar(&labs, "labs", "", "Lab folders to add linter tests to (space separated)")

	if err := fs.Parse(args); err != nil {
		exitErr(err, "Error parsing flags")
	}
	for dir := range strings.SplitSeq(labs, " ") {
		fmt.Printf("Updating %q in %s\n", lintFile, courseRepoPath(dir))
		if err := generateGoFromTemplate(dir, lintFile, lintTmplFile, linterTmplFS); err != nil {
			exitErr(err, "Error generating linter test")
		}
	}
}

type GoTemplateConfig struct {
	Package string
	Lab     string
}

//go:embed linter_qf_test.tmpl
var linterTmplFS embed.FS

const (
	lintTmplFile = "linter_qf_test.tmpl"
	lintFile     = "linter_qf_test.go"
)

func firstPathElem(path string) string {
	i := strings.Index(path, string(os.PathSeparator))
	if i < 0 {
		return path
	}
	return path[0:i]
}

func generateGoFromTemplate(dir, file, tmplFile string, tmplFS fs.FS) error {
	path := filepath.Join(dir, file)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	tmpl, err := template.ParseFS(tmplFS, tmplFile)
	if err != nil {
		return err
	}
	config := &GoTemplateConfig{
		Package: packageName(dir),
		Lab:     firstPathElem(dir),
	}
	return tmpl.Execute(f, config)
}

// packageName returns the package name of the first Go file in the given directory.
// If no Go files are found, the base directory name is returned.
func packageName(dir string) string {
	noPkg := filepath.Base(dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return noPkg
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".go" {
			filePath := filepath.Join(dir, entry.Name())
			fset := token.NewFileSet()
			// parse only the package clause of the file
			f, err := parser.ParseFile(fset, filePath, nil, parser.PackageClauseOnly)
			if err != nil {
				// if we can't parse the file, try the next one
				continue
			}
			return f.Name.Name
		}
	}
	return noPkg
}

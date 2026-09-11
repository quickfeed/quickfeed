package ci

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

func TestCopyDir(t *testing.T) {
	src := t.TempDir()
	files := map[string]string{
		filepath.Join("lab1", "lab1.go"):                "package lab1\n",
		filepath.Join("lab1", "internal", "helper.go"):  "package internal\n",
		filepath.Join("lab2", "lab2.go"):                "package lab2\n",
		"go.mod":                                        "module example\n",
		filepath.Join(".git", "config"):                 "[core]\n",
		filepath.Join(".git", "refs", "heads", "main"):  "deadbeef\n",
		filepath.Join("lab1", ".git", "unexpected.txt"): "nested git dir\n",
	}
	for name, content := range files {
		path := filepath.Join(src, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// An executable file must stay executable in the copy.
	script := filepath.Join(src, "lab1", "run.sh")
	if err := os.WriteFile(script, []byte("#!/bin/bash\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(t.TempDir(), "submitted")
	if err := copyDir(src, dst); err != nil {
		t.Fatal(err)
	}

	wantFiles := []string{
		"go.mod",
		filepath.Join("lab1", "internal", "helper.go"),
		filepath.Join("lab1", "lab1.go"),
		filepath.Join("lab1", "run.sh"),
		filepath.Join("lab2", "lab2.go"),
	}
	var gotFiles []string
	if err := filepath.WalkDir(dst, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dst, path)
		if err != nil {
			return err
		}
		gotFiles = append(gotFiles, rel)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	slices.Sort(gotFiles)
	if !slices.Equal(wantFiles, gotFiles) {
		t.Errorf("copyDir() copied %q, want %q", gotFiles, wantFiles)
	}

	for _, name := range wantFiles {
		want, ok := files[name]
		if !ok {
			continue // the run.sh script is not in the files map
		}
		got, err := os.ReadFile(filepath.Join(dst, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Errorf("copyDir() content of %q = %q, want %q", name, got, want)
		}
	}

	if runtime.GOOS != "windows" {
		// Windows does not carry the Unix permission bits.
		info, err := os.Stat(filepath.Join(dst, "lab1", "run.sh"))
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != 0o755 {
			t.Errorf("copyDir() mode of lab1/run.sh = %o, want %o", got, 0o755)
		}
	}
}

func TestCopyDirMissingSource(t *testing.T) {
	src := filepath.Join(t.TempDir(), "missing")
	if err := copyDir(src, t.TempDir()); err == nil {
		t.Error("copyDir() error = nil, want an error for a missing source directory")
	}
}

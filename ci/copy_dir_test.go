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

// TestCopyDirOverlappingPaths checks that a copy that would descend into its
// own destination, or copy the destination into itself, is refused. RunTests
// creates its destination under os.TempDir(), so a submission directory that
// names a temporary directory is all it takes.
func TestCopyDirOverlappingPaths(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "submission")
	if err := os.MkdirAll(filepath.Join(src, "lab1"), 0o700); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		src  string
		dst  string
	}{
		{name: "SameDirectory", src: src, dst: src},
		{name: "DestinationInsideSource", src: src, dst: filepath.Join(src, quickfeedTestsPath, "user-labs")},
		{name: "DestinationIsExistingSubdirectory", src: src, dst: filepath.Join(src, "lab1")},
		{name: "SourceInsideDestination", src: src, dst: root},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := copyDir(tc.src, tc.dst); err == nil {
				t.Errorf("copyDir(%q, %q) error = nil, want an overlap error", tc.src, tc.dst)
			}
		})
	}
}

// TestCopyDirSiblingPaths checks that the overlap detection does not reject a
// destination whose path merely starts with the source's path.
func TestCopyDirSiblingPaths(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "lab")
	if err := os.MkdirAll(src, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "lab.go"), []byte("package lab\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "lab-copy")
	if err := copyDir(src, dst); err != nil {
		t.Errorf("copyDir(%q, %q) error = %v, want nil", src, dst, err)
	}
	if _, err := os.Stat(filepath.Join(dst, "lab.go")); err != nil {
		t.Error(err)
	}
}

// TestCopyDirSymlinkedPaths checks that overlapping paths are recognized
// through a symbolic link, as on macOS, where /tmp links to /private/tmp: the
// two paths name the same directory only once they are resolved.
func TestCopyDirSymlinkedPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating a symbolic link on Windows requires elevated privileges")
	}
	root := t.TempDir()
	src := filepath.Join(root, "submission")
	if err := os.MkdirAll(src, 0o700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(src, link); err != nil {
		t.Fatal(err)
	}
	// The destination is created under the link, and thereby inside the source.
	dst := filepath.Join(link, quickfeedTestsPath)
	if err := copyDir(src, dst); err == nil {
		t.Errorf("copyDir(%q, %q) error = nil, want an overlap error through the link", src, dst)
	}
	// The same, with the roles of the link and the resolved path exchanged.
	dst = filepath.Join(src, quickfeedTestsPath)
	if err := copyDir(link, dst); err == nil {
		t.Errorf("copyDir(%q, %q) error = nil, want an overlap error through the link", link, dst)
	}
}

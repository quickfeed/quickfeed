package fileop_test

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/quickfeed/quickfeed/internal/fileop"
)

// tempPrefix stands in for the temporary directory a test run is copied into.
const tempPrefix = "quickfeed-tests"

func TestCopyDir(t *testing.T) {
	src := t.TempDir()
	files := map[string]string{
		filepath.Join("lab1", "lab1.go"):                "package lab1\n",
		filepath.Join("lab1", "internal", "helper.go"):  "package internal\n",
		filepath.Join("lab2", "lab2.go"):                "package lab2\n",
		"go.mod":                                        "module example\n",
		".gitignore":                                    "bin/\n",
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
	// An empty directory is part of the tree and must be copied too.
	if err := os.Mkdir(filepath.Join(src, "lab3"), 0o700); err != nil {
		t.Fatal(err)
	}

	// The destination's parent does not exist yet; CopyDir must create it.
	dst := filepath.Join(t.TempDir(), tempPrefix, "submitted")
	if err := fileop.CopyDir(src, dst); err != nil {
		t.Fatal(err)
	}

	// Only .git directories are skipped; a .gitignore file is a regular file.
	wantFiles := []string{
		".gitignore",
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
		t.Errorf("CopyDir() copied %q, want %q", gotFiles, wantFiles)
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
			t.Errorf("CopyDir() content of %q = %q, want %q", name, got, want)
		}
	}

	if info, err := os.Stat(filepath.Join(dst, "lab3")); err != nil || !info.IsDir() {
		t.Errorf("CopyDir() did not copy the empty directory lab3: %v", err)
	}

	if runtime.GOOS != "windows" {
		// Windows does not carry the Unix permission bits.
		info, err := os.Stat(filepath.Join(dst, "lab1", "run.sh"))
		if err != nil {
			t.Fatal(err)
		}
		if got := info.Mode().Perm(); got != 0o755 {
			t.Errorf("CopyDir() mode of lab1/run.sh = %o, want %o", got, 0o755)
		}
	}
}

func TestCopyDirMissingSource(t *testing.T) {
	src := filepath.Join(t.TempDir(), "missing")
	if err := fileop.CopyDir(src, t.TempDir()); err == nil {
		t.Error("CopyDir() error = nil, want an error for a missing source directory")
	}
}

// TestCopyDirOverlappingPaths checks that a copy that would descend into its
// own destination, or copy the destination into itself, is refused. A test run
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
		{name: "DestinationInsideSource", src: src, dst: filepath.Join(src, tempPrefix, "user-labs")},
		{name: "DestinationIsExistingSubdirectory", src: src, dst: filepath.Join(src, "lab1")},
		{name: "SourceInsideDestination", src: src, dst: root},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := fileop.CopyDir(tc.src, tc.dst); err == nil {
				t.Errorf("CopyDir(%q, %q) error = nil, want an overlap error", tc.src, tc.dst)
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
	if err := fileop.CopyDir(src, dst); err != nil {
		t.Errorf("CopyDir(%q, %q) error = %v, want nil", src, dst, err)
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
	dst := filepath.Join(link, tempPrefix)
	if err := fileop.CopyDir(src, dst); err == nil {
		t.Errorf("CopyDir(%q, %q) error = nil, want an overlap error through the link", src, dst)
	}
	// The same, with the roles of the link and the resolved path exchanged.
	dst = filepath.Join(src, tempPrefix)
	if err := fileop.CopyDir(link, dst); err == nil {
		t.Errorf("CopyDir(%q, %q) error = nil, want an overlap error through the link", link, dst)
	}
}

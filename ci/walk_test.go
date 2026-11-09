package ci

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalkSkipsSymlinks(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "real"), 0o755); err != nil {
		t.Fatal(err)
	}
	regular := filepath.Join(dir, "real", "file.txt")
	if err := os.WriteFile(regular, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real", filepath.Join(dir, "dirlink")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("missing", filepath.Join(dir, "danglinglink")); err != nil {
		t.Fatal(err)
	}

	files, err := walk(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || string(files[regular]) != "content" {
		t.Errorf("walk() = %v, want only %s", files, regular)
	}
}

package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
)

func TestSetRoot(t *testing.T) {
	// assume working directory is within the quickfeed repository
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()
	setRoot()
}

func TestSetRootWorkingDirOutside(t *testing.T) {
	// reset quickfeedRoot to empty string; to circumvent the init() function
	quickfeedRoot = ""

	// set working directory to be outside the quickfeed repository
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chdir("/")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()

	// setRoot() is expected to panic; this test will fail if it does not panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected setRoot() to panic")
		}
	}()
	setRoot()
}

func TestSetRootWrongGitRepo(t *testing.T) {
	// reset quickfeedRoot to empty string; to circumvent the init() function
	quickfeedRoot = ""

	// set working directory to be inside a git repository that is not the quickfeed repository
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	gitRepo := t.TempDir()
	err = os.Chdir(gitRepo)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()

	// create a git repository
	_, err = git.PlainInit(gitRepo, false)
	if err != nil {
		t.Fatal(err)
	}

	// create go.mod file with a different module name than the quickfeed module
	modFile := filepath.Join(gitRepo, "go.mod")
	f, err := os.Create(modFile)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString("module github.com/wrong/module\n"); err != nil {
		t.Fatal(err)
	}

	// setRoot() is expected to panic; this test will fail if it does not panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected setRoot() to panic")
		}
	}()
	setRoot()
}

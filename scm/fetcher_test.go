package scm_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

func TestClone(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	s := scm.GetTestSCM(t)
	userName, err := s.GetUserName(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	dstDir := t.TempDir()
	assignmentDir, err := s.Clone(ctx, &scm.CloneOptions{
		Organization: qfTestOrg,
		Repository:   qf.StudentRepoName(userName),
		DestDir:      dstDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	testsDir, err := s.Clone(ctx, &scm.CloneOptions{
		Organization: qfTestOrg,
		Repository:   qf.TestsRepo,
		DestDir:      dstDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Note: the following depends on the actual content of
	// the <student>-labs and tests repositories of the qfTestOrg.
	found, err := exists(filepath.Join(assignmentDir, "lab1"))
	if !found {
		t.Fatalf("lab1 not found in 'assignments': %v", err)
	}
	found, err = exists(filepath.Join(assignmentDir, "lab2"))
	if found {
		t.Fatalf("lab2 found in 'assignments' unexpectedly: %v", err)
	}
	expectedTestsDirs := []string{
		"lab1",
		"lab2",
		"lab3",
		"lab4",
		"lab5",
		"lab6",
		"scripts",
	}
	for _, dir := range expectedTestsDirs {
		found, err = exists(filepath.Join(testsDir, dir))
		if !found {
			t.Errorf("%s not found in 'tests': %v", dir, err)
		}
	}
}

func TestCloneBranch(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	s := scm.GetTestSCM(t)
	userName, err := s.GetUserName(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	dstDir := t.TempDir()
	assignmentDir, err := s.Clone(ctx, &scm.CloneOptions{
		Organization: qfTestOrg,
		Repository:   qf.StudentRepoName(userName),
		DestDir:      dstDir,
		Branch:       "hotfix",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Note: the following depends on the actual existence of the lab2
	// folder in the <student>-labs repository of the qfTestOrg.
	found, err := exists(filepath.Join(assignmentDir, "lab2"))
	if !found {
		t.Fatalf("lab2 not found in %s: %v", assignmentDir, err)
	}
}

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

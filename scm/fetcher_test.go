package scm_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/fileop"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/kit/sh"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

const debug = false

func mustRun(wd, cmd string, args ...string) {
	c := exec.Command(cmd, args...)
	c.Dir = wd
	if debug {
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		log.Println("running:", cmd, strings.Join(args, " "))
	}
	if err := c.Run(); err != nil {
		panic(fmt.Sprintf("failed to run %s %v: %v", cmd, args, err))
	}
}

// prepareGitRepo creates copies src/repo folder to dst and initializes
// dst/repo as a git repository and adds a single file lab1/lab1.go.
func prepareGitRepo(src, dst, repo string) error {
	if err := fileop.CopyDir(filepath.Join(src, repo), dst); err != nil {
		return err
	}
	gitRepo := filepath.Join(dst, repo)
	r, err := git.PlainInit(gitRepo, false)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	_, err = w.Add("lab1")
	if err != nil {
		return err
	}
	_, err = w.Commit("added lab1", &git.CommitOptions{})
	if err != nil {
		return err
	}
	return nil
}

func TestFileClone(t *testing.T) {
	repoPath := t.TempDir()
	t.Setenv("QUICKFEED_REPOSITORY_PATH", repoPath)

	src := filepath.Join(env.TestdataPath(), qtest.MockOrg)
	dst := filepath.Join(repoPath, qtest.MockOrg)
	err := prepareGitRepo(src, dst, qf.AssignmentsRepo)
	if err != nil {
		t.Fatal(err)
	}

	s := scm.NewMockedGithubSCMClient(qtest.Logger(t))

	dstDir := t.TempDir()
	assignmentDir, err := s.Clone(context.Background(), &scm.CloneOptions{
		Organization: qtest.MockOrg,
		Repository:   qf.AssignmentsRepo,
		DestDir:      dstDir,
	})
	if err != nil {
		t.Fatal(err)
	}

	// The following depends on the actual content of
	// the testdata/courses/qf102-2022/assignments folder.
	found, err := exists(filepath.Join(assignmentDir, "lab1"))
	if !found {
		t.Fatalf("lab1 not found in %s: %v", assignmentDir, err)
	}
}

func TestClone(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	s, userName := scm.GetTestSCM(t)

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

func appendToFile(filename, text string) (err error) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := f.Close()
		if err == nil {
			err = closeErr
		}
	}()
	_, err = f.WriteString(text)
	return
}

// Test that we can clone a repository, update it (commit and push) and clone it again twice.
// The two last clones are in a different directory.
// The third clone is actually a fast-forward pull.
func TestCloneTwice(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	s, _ := scm.GetTestSCM(t)

	ctx := context.Background()
	dstDir := t.TempDir()

	testsDir, err := s.Clone(ctx, &scm.CloneOptions{
		Organization: qfTestOrg,
		Repository:   qf.TestsRepo,
		DestDir:      dstDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if found, err := exists(testsDir); !found {
		t.Fatalf("%s not found: %v", testsDir, err)
	}
	twiceMsg := fmt.Sprintf("Update tests repo %s\n", time.Now().Format(time.Kitchen))
	if err := appendToFile(filepath.Join(testsDir, "README.md"), twiceMsg); err != nil {
		t.Fatal(err)
	}
	commitMsg := fmt.Sprintf("Clone twice commit %s", time.Now().Format(time.Kitchen))
	if err := sh.RunA("git", "-C", testsDir, "commit", "-a", "-m", commitMsg); err != nil {
		t.Fatal(err)
	}
	if err := sh.RunA("git", "-C", testsDir, "push"); err != nil {
		t.Fatal(err)
	}

	// Clone to a new directory to ensure that we get a new clone with the change we just made.
	dstDir = t.TempDir()

	testsDir, err = s.Clone(ctx, &scm.CloneOptions{
		Organization: qfTestOrg,
		Repository:   qf.TestsRepo,
		DestDir:      dstDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(testsDir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), twiceMsg[:len(twiceMsg)-2]) {
		t.Fatalf("README.md does not contain %q", twiceMsg)
	}

	// Clone to the same directory to test that we get a fast-forward pull.
	testsDir, err = s.Clone(ctx, &scm.CloneOptions{
		Organization: qfTestOrg,
		Repository:   qf.TestsRepo,
		DestDir:      dstDir,
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err = os.ReadFile(filepath.Join(testsDir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), twiceMsg[:len(twiceMsg)-2]) {
		t.Fatalf("README.md does not contain %q", twiceMsg)
	}
}

func TestCloneBranch(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	s, userName := scm.GetTestSCM(t)

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

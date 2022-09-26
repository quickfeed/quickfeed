package ci

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/quickfeed/quickfeed/scm"
)

type RepoInfo struct {
	Repo   string
	Branch string
}

type CloneInfo struct {
	CourseCode        string
	JobOwner          string
	OrganizationPath  string
	CurrentAssignment string
	DestDir           string
	CloneRepos        []RepoInfo
}

// CloneRepositories clones the repositories for the given course organization.
func CloneRepositories(ctx context.Context, sc scm.SCM, info *CloneInfo) ([]string, error) {
	defer timer(info.JobOwner, info.CourseCode, cloneTimeGauge)()

	cloneDirs := make([]string, 0, len(info.CloneRepos))
	for _, repo := range info.CloneRepos {
		dir, err := sc.Clone(ctx, &scm.CloneOptions{
			Organization: info.OrganizationPath,
			Repository:   repo.Repo,
			DestDir:      info.DestDir,
			Branch:       repo.Branch,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to clone %q repository: %w", repo.Repo, err)
		}
		if hasAssignment(dir, info.CurrentAssignment) != nil {
			return nil, err
		}
		cloneDirs = append(cloneDirs, dir)
	}
	return cloneDirs, nil
}

// hasAssignment return nil if the assignment exists in the given clone dir.
func hasAssignment(cloneDir, currentAssignment string) error {
	// Check that there is code for the current assignment in clone dir
	if ok, err := exists(filepath.Join(cloneDir, currentAssignment)); !ok {
		return fmt.Errorf("%s directory does not contain %q: %w", cloneDir, currentAssignment, err)
	}
	return nil
}

// ScanStudentRepo returns an error if the student's repository contains the session secret environment variable.
// Note: This scan may be costly for large repositories.
func ScanStudentRepo(submittedDir, course, jobOwner string) error {
	defer timer(jobOwner, course, validationTimeGauge)()

	// Walk the student's submitted code directory
	files, err := walk(submittedDir)
	if err != nil {
		return err
	}
	// Ensure that the student code files does not contain the session secret environment variable.
	for file, content := range files {
		if strings.Contains(string(content), secretEnvName) {
			return fmt.Errorf("file %q in (%s/%s) contains the %q environment variable", filepath.Base(file), course, jobOwner, secretEnvName)
		}
		// We could add more checks here.
	}
	return nil
}

// exists returns true if path exists and is a directory.
// Otherwise, it returns false and an error.
func exists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil && fi.IsDir() {
		return true, nil
	}
	return false, err
}

// walk walks the student code repository and returns a map of file names and their contents.
func walk(dir string) (map[string][]byte, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, err
	}
	files := make(map[string][]byte)
	err := filepath.WalkDir(dir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			// Walk unable to read path; stop walking the tree
			return err
		}
		if !info.IsDir() {
			if files[path], err = os.ReadFile(path); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

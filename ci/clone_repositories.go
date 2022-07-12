package ci

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"go.uber.org/zap"
)

func (r RunData) cloneRepositories(ctx context.Context, logger *zap.SugaredLogger, dstDir string) error {
	logger.Debugf("Cloning repositories for %s", r)

	// TODO(meling): Update this for GitHub web app.
	// The scm client should ideally be passed in instead of creating another instance.
	sc, err := scm.NewSCMClient(logger, r.Course.GetAccessToken())
	if err != nil {
		return fmt.Errorf("failed to create SCM Client: %w", err)
	}

	start := time.Now()
	testsDir, err := sc.Clone(ctx, &scm.CloneOptions{
		Organization: r.Course.GetOrganizationPath(),
		Repository:   qf.TestsRepo,
		DestDir:      dstDir,
	})
	if err != nil {
		return fmt.Errorf("failed to clone %q repository: %w", qf.TestsRepo, err)
	}

	assignmentsDir, err := sc.Clone(ctx, &scm.CloneOptions{
		Organization: r.Course.GetOrganizationPath(),
		Repository:   r.Repo.Name(),
		DestDir:      dstDir,
		Branch:       r.BranchName,
	})
	if err != nil {
		return fmt.Errorf("failed to clone %q repository: %w", qf.AssignmentRepo, err)
	}
	logger.Debugf("Cloning time:    %v", time.Since(start))
	start = time.Now()
	defer func() {
		logger.Debugf("Validation time: %v", time.Since(start))
	}()
	return r.validate(testsDir, assignmentsDir)
}

// validate performs various checks on the cloned repositories.
func (r RunData) validate(testsDir, assignmentsDir string) error {
	// Check that there are tests for the current assignment {{ .AssignmentName }}
	if ok, err := exists(filepath.Join(testsDir, r.Assignment.GetName())); !ok {
		return fmt.Errorf("tests directory does not contain %q: %w", r.Assignment.GetName(), err)
	}
	// Check that there is student code directory for the current assignment {{ .AssignmentName }}
	if ok, err := exists(filepath.Join(assignmentsDir, r.Assignment.GetName())); !ok {
		return fmt.Errorf("assignments directory does not contain %q: %w", r.Assignment.GetName(), err)
	}

	// Note: The following check may be costly if the student code is large.
	// Hence, we may consider adding a flag to skip this check. A flag could
	// be added to qf.Assignment or qf.Course, both accessible via RunData.

	// Walk the student's assignments directory
	files, err := walk(assignmentsDir)
	if err != nil {
		return err
	}
	// Ensure that the student code files does not contain the session secret environment variable.
	for file, content := range files {
		if strings.Contains(string(content), secretEnvName) {
			return fmt.Errorf("file %q in %s contains %s environment variable", filepath.Base(file), r, secretEnvName)
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

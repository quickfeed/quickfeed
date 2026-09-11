package main

import (
	"errors"
	"flag"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/quickfeed/quickfeed/internal/env"
	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

// commonFlags holds the flags that every qcm subcommand accepts.
type commonFlags struct {
	org     string
	code    string
	dir     string
	token   string
	verbose bool
}

func (c *commonFlags) register(fs *flag.FlagSet) {
	fs.StringVar(&c.org, "course", "", "the course's GitHub `organization`, e.g., dat320-2025 (required)")
	fs.StringVar(&c.code, "code", "", "the course `code`; derived from -course if not given")
	fs.StringVar(&c.dir, "dir", env.RepositoryPath(), "root `directory` holding the course repositories")
	fs.StringVar(&c.token, "token", os.Getenv("GITHUB_ACCESS_TOKEN"), "GitHub personal access `token`; defaults to $GITHUB_ACCESS_TOKEN")
	fs.BoolVar(&c.verbose, "v", false, "print debug logging to stderr")
}

// validate checks the common flags and fills in the derived values. It must be
// called before any other method.
func (c *commonFlags) validate() error {
	if c.org == "" {
		return errors.New("missing required flag -course")
	}
	if c.code == "" {
		c.code = courseCode(c.org)
	}
	if c.dir == "" {
		return errors.New("flag -dir must name a directory")
	}
	// The ci package resolves a course's clone directory from the repository
	// path in the environment; point it at -dir so that both agree on where
	// the course repositories are.
	os.Setenv("QUICKFEED_REPOSITORY_PATH", c.dir)
	return nil
}

// course returns the course described by the common flags. Its CloneDir is
// <dir>/<org>, holding the tests and assignments repositories.
func (c *commonFlags) course() *qf.Course {
	return &qf.Course{Code: c.code, ScmOrganizationName: c.org}
}

// testsDir returns the path to the local clone of the tests repository.
func (c *commonFlags) testsDir() string {
	return filepath.Join(c.course().CloneDir(), qf.TestsRepo)
}

// assignmentsDir returns the path to the local clone of the assignments repository.
func (c *commonFlags) assignmentsDir() string {
	return filepath.Join(c.course().CloneDir(), qf.AssignmentsRepo)
}

// logger returns the logger for the command. Without -v, only warnings and
// errors are printed, since the server's debug records would otherwise bury
// the command's own output.
func (c *commonFlags) logger(stderr io.Writer) *slog.Logger {
	level := slog.LevelWarn
	if c.verbose {
		level = slog.LevelDebug
	}
	return qlog.NewLevel(stderr, level)
}

// scmClient returns an SCM client for the course's organization.
func (c *commonFlags) scmClient(logger *slog.Logger) (scm.SCM, error) {
	if c.token == "" {
		return nil, errors.New("missing GitHub access token; set -token or the GITHUB_ACCESS_TOKEN environment variable")
	}
	return scm.NewSCMClient(logger, c.token)
}

// yearSuffix matches the four-digit year that a course organization is
// conventionally suffixed with, e.g., the -2025 in dat320-2025.
var yearSuffix = regexp.MustCompile(`-[0-9]{4}$`)

// courseCode derives a course code from the course's GitHub organization by
// stripping a trailing four-digit year, if present, and upper-casing the rest:
// dat320-2025 becomes DAT320, whereas my-course becomes MY-COURSE.
//
// The code must match the image named by the course's run scripts when the
// course builds its own Docker image, that is, #image/<lowercase code>.
func courseCode(org string) string {
	return strings.ToUpper(yearSuffix.ReplaceAllString(org, ""))
}

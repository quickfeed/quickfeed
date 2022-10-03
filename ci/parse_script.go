package ci

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/quickfeed/quickfeed/qf"
)

// parseTestRunnerScript returns a job specifying the docker image and commands
// to be executed by the docker image. The job's commands are extracted from
// the test runner script (run.sh) associated with the RunData's assignment.
//
// The script may use the following environment variables:
//
//	TESTS       - to access the tests (cloned from the course's tests repository)
//	ASSIGNMENTS - to access the assignments (cloned from the course's assignments repository)
//	SUBMITTED   - to access the student's or group's submitted code (cloned from the student/group repository)
//	CURRENT     - name of the current assignment folder
//	QUICKFEED_SESSION_SECRET - typically used by the test code; not the script itself
func (r RunData) parseTestRunnerScript(secret, destDir string) (*Job, error) {
	image, commands, err := parseRunScript(r.Assignment.GetRunScriptContent())
	if err != nil {
		return nil, fmt.Errorf("failed to parse run script for assignment %s in %s: %w", r.Assignment.GetName(), r.Repo.GetTestURL(), err)
	}
	if r.EnvVarsFn == nil {
		// For docker runs, the home path is set to QuickFeedPath = /quickfeed
		r.EnvVarsFn = func(secret, _ string) []string {
			// QuickFeedPath is the home path (inside the container) bound to the temporary tests directory
			return EnvVars(secret, QuickFeedPath, r.Repo.Name(), r.Assignment.GetName())
		}
	}
	return &Job{
		Name:       r.String(),
		Image:      image,
		Dockerfile: r.Course.Dockerfile,
		BindDir:    destDir,
		Env:        r.EnvVarsFn(secret, destDir),
		Commands:   commands,
	}, nil
}

func parseRunScript(scriptContent string) (image string, commands []string, err error) {
	s := strings.Split(scriptContent, "\n")
	if len(s) < 2 {
		return "", nil, errors.New("empty run script")
	}
	parts := strings.Split(s[0], "#image/")
	if len(parts) < 2 {
		return "", nil, errors.New("no docker image specified in run script")
	}
	return strings.ToLower(parts[1]), s[1:], nil
}

func EnvVars(sessionSecret, home, repoName, currentAssignment string) []string {
	envMap := map[string]string{
		"HOME":        home,
		"TESTS":       filepath.Join(home, qf.TestsRepo),
		"ASSIGNMENTS": filepath.Join(home, qf.AssignmentRepo),
		"SUBMITTED":   filepath.Join(home, repoName),
		"CURRENT":     currentAssignment,
		secretEnvName: sessionSecret,
	}
	envVars := make([]string, 0, len(envMap))
	for varName, value := range envMap {
		envVars = append(envVars, fmt.Sprintf("%s=%s", varName, value))
	}
	return envVars
}

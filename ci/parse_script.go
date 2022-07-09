package ci

import (
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
//   TESTS - to access the tests (cloned from the tests repository)
//   ASSIGNMENTS - to access the assignments (cloned from the assignments repository)
//   CURRENT - name of the current assignment folder
//   QUICKFEED_SESSION_SECRET - typically used by the test code; not the script itself
func (r RunData) parseTestRunnerScript(secret string) (*Job, error) {
	s := strings.Split(r.Assignment.GetScriptFile(), "\n")
	if len(s) < 2 {
		return nil, fmt.Errorf("no run script for assignment %s in %s", r.Assignment.GetName(), r.Repo.GetTestURL())
	}
	parts := strings.Split(s[0], "#image/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("no docker image specified in run script for assignment %s in %s", r.Assignment.GetName(), r.Repo.GetTestURL())
	}
	return &Job{
		Name:     r.String(),
		Image:    parts[1],
		Env:      r.envVars(secret),
		Commands: s[1:],
	}, nil
}

func (r RunData) envVars(sessionSecret string) []string {
	envMap := map[string]string{
		"TESTS":       filepath.Join(QuickFeedPath, qf.TestsRepo),
		"ASSIGNMENTS": filepath.Join(QuickFeedPath, qf.AssignmentRepo),
		"CURRENT":     r.Assignment.GetName(),
		secretEnvName: sessionSecret,
	}
	envVars := make([]string, 0, len(envMap))
	for varName, value := range envMap {
		envVars = append(envVars, fmt.Sprintf("%s=%s", varName, value))
	}
	return envVars
}

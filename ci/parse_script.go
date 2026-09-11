package ci

import (
	"errors"
	"fmt"
	"os"
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
func (r *RunData) parseTestRunnerScript(secret, destDir string) (*Job, error) {
	scriptContent, err := r.loadRunScript()
	if err != nil {
		return nil, err
	}
	image, language, commands, err := ParseRunScript(scriptContent)
	if err != nil {
		return nil, fmt.Errorf("parsing run script for assignment %s in %s: %w", r.Assignment.GetName(), r.Repo.GetTestURL(), err)
	}
	if r.EnvVarsFn == nil {
		// For docker runs, the home path is set to QuickFeedPath = /quickfeed
		r.EnvVarsFn = func(secret, _ string) []string {
			// QuickFeedPath is the home path (inside the container) bound to the temporary tests directory
			vars := EnvVars(secret, QuickFeedPath, r.Repo.Name(), r.Assignment.GetName())
			if cfg, ok := languages[language]; ok {
				vars = append(vars, cfg.envVars...)
			}
			return vars
		}
	}
	testsDir := filepath.Join(r.Course.CloneDir(), qf.TestsRepo)
	assignmentDir := filepath.Join(r.Course.CloneDir(), qf.AssignmentsRepo)
	return &Job{
		Name:     r.String(),
		Image:    image,
		Language: language,
		BindDir:  destDir,
		ReadOnlyMounts: map[string]string{
			testsDir:      filepath.Join(QuickFeedPath, qf.TestsRepo),
			assignmentDir: filepath.Join(QuickFeedPath, qf.AssignmentsRepo),
		},
		Env: r.EnvVarsFn(secret, destDir),
		// The build check runs before the course's run script, so that a
		// compilation failure can be attributed to the submitted code.
		Commands: append(buildCheckCommands(language), commands...),
	}, nil
}

func (r *RunData) loadRunScript() (string, error) {
	const (
		scriptFile   = "run.sh"
		scriptFolder = "scripts"
	)
	courseTestsDir := filepath.Join(r.Course.CloneDir(), qf.TestsRepo)
	runScript := filepath.Join(courseTestsDir, r.Assignment.GetName(), scriptFile)
	if _, err := os.Stat(runScript); os.IsNotExist(err) {
		// If the assignment does not have a run.sh script, use the default run.sh script
		runScript = filepath.Join(courseTestsDir, scriptFolder, scriptFile)
		if _, err := os.Stat(runScript); os.IsNotExist(err) {
			return "", fmt.Errorf("run script not found for %s: %w", r.Course.GetCode(), err)
		}
	}
	b, err := os.ReadFile(runScript)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ParseRunScript parses the content of a course's test runner script and
// returns the Docker image named by the script's mandatory #image/ directive,
// the programming language named by its optional #language/ directive, and the
// remaining lines as the commands to run.
//
// The script must name an image and must hold at least one command; a script
// consisting of directives and blank lines alone cannot produce test results,
// and is reported here rather than as a mystifying empty test run. Teachers can
// therefore check their run scripts with this function; see cmd/qcm.
func ParseRunScript(scriptContent string) (image, language string, commands []string, err error) {
	lines := strings.Split(scriptContent, "\n")
	if len(lines) < 3 {
		return "", "", nil, errors.New("empty run script")
	}
	parts := strings.Split(lines[0], "#image/")
	if len(parts) < 2 {
		return "", "", nil, errors.New("no docker image specified in run script")
	}
	image = strings.ToLower(parts[1])
	hasCommand := false
	for _, line := range lines[1:] {
		if lang, found := strings.CutPrefix(line, "#language/"); found {
			language = strings.ToLower(strings.TrimSpace(lang))
			continue
		}
		hasCommand = hasCommand || strings.TrimSpace(line) != ""
		commands = append(commands, line)
	}
	if !hasCommand {
		return "", "", nil, errors.New("no commands in run script")
	}
	return image, language, commands, nil
}

func EnvVars(sessionSecret, home, repoName, currentAssignment string) []string {
	envMap := map[string]string{
		"HOME":        home,
		"TESTS":       filepath.Join(home, qf.TestsRepo),
		"ASSIGNMENTS": filepath.Join(home, qf.AssignmentsRepo),
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

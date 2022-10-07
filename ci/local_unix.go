//go:build linux || darwin

package ci

import (
	"context"
	"os/exec"
	"strings"
)

// Local is an implementation of the CI interface executing code locally.
type Local struct{}

// Run implements the CI interface. This method blocks until the job has been
// completed or an error occurs, e.g., the context times out.
func (l *Local) Run(ctx context.Context, job *Job) (string, error) {
	// TODO: Execute tests in something like: os.CreateTemp(os.TempDir(), "local-ci")
	cmd := exec.Command("/bin/bash", "-c", strings.Join(job.Commands, "\n"))
	cmd.Env = job.Env
	b, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

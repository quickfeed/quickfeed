package ci

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// Local is an implementation of the CI interface executing code locally.
type Local struct{}

// Run implements the CI interface. This method blocks until the job has been
// completed or an error occurs, e.g., the context times out.
func (l *Local) Run(ctx context.Context, job *Job, user string, timeout time.Duration) (string, error) {
	cmd := exec.Command("bash", "-c", strings.Join(job.Commands, "\n"))
	b, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(b), nil
	// return "", errors.New("no local implementation of CI for windows")
}

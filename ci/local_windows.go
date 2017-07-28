package ci

import (
	"context"
	"errors"
)

// Local is an implementation of the CI interface executing code locally.
type Local struct{}

// Run implements the CI interface. This method blocks until the job has been
// completed or an error occurs, e.g., the context times out.
func (l *Local) Run(ctx context.Context, job Job) (string, error) {
	return "", errors.New("no local implementation of CI for windows")
}

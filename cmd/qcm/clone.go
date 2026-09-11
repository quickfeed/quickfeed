package main

import (
	"context"
	"fmt"
	"io"

	"github.com/quickfeed/quickfeed/internal/qlog"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

const cloneSynopsis = `Usage: qcm clone -course ORG [flags]

Clone the course's tests and assignments repositories into <dir>/<org>.
An existing clone is updated with git pull instead of being cloned again.`

func cloneCmd(args []string, stdout, stderr io.Writer) error {
	fs := newFlagSet("clone", stderr, cloneSynopsis)
	var c commonFlags
	c.register(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := c.validate(); err != nil {
		return err
	}

	logger := c.logger(stderr)
	client, err := c.scmClient(logger)
	if err != nil {
		return err
	}
	ctx := qlog.NewContext(context.Background(), logger)
	course := c.course()
	for _, repo := range []string{qf.TestsRepo, qf.AssignmentsRepo} {
		path, err := client.Clone(ctx, &scm.CloneOptions{
			Organization: c.org,
			Repository:   repo,
			DestDir:      course.CloneDir(),
		})
		if err != nil {
			return fmt.Errorf("cloning %s/%s repository: %w", c.org, repo, err)
		}
		fmt.Fprintf(stdout, "%s repository: %s\n", repo, path)
	}
	return nil
}

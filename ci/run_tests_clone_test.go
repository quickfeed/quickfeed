package ci

import (
	"context"
	"strings"
	"testing"

	"github.com/quickfeed/quickfeed/internal/rand"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

func TestCloneAndCopyRunTests(t *testing.T) {
	t.Setenv("QUICKFEED_REPOSITORY_PATH", "$HOME/tmp/courses")
	qfTestOrg := scm.GetTestOrganization(t)
	sc, qfUserName := scm.GetTestSCM(t)

	dstDir := t.TempDir()

	course := &qf.Course{
		Code:                "QF101",
		ScmOrganizationName: qfTestOrg,
	}
	repo := qf.RepoURL{ProviderURL: "github.com", Organization: qfTestOrg}
	runData := &RunData{
		Course: course,
		Assignment: &qf.Assignment{
			Name: "lab1",
		},
		Repo: &qf.Repository{
			HTMLURL:  repo.StudentRepoURL(qfUserName),
			RepoType: qf.Repository_USER,
		},
		JobOwner: "muggles",
		CommitID: rand.String()[:7],
	}

	ctx := context.Background()
	clonedAssignmentsRepo, err := sc.Clone(ctx, &scm.CloneOptions{
		Organization: course.GetScmOrganizationName(),
		Repository:   qf.AssignmentsRepo,
		DestDir:      course.CloneDir(),
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(clonedAssignmentsRepo)

	clonedTestsRepo, err := sc.Clone(ctx, &scm.CloneOptions{
		Organization: course.GetScmOrganizationName(),
		Repository:   qf.TestsRepo,
		DestDir:      course.CloneDir(),
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(clonedTestsRepo)

	if err := runData.clone(ctx, sc, dstDir); err != nil {
		t.Error(err)
	}
	localRunner := Local{}
	out, err := localRunner.Run(ctx, &Job{
		Commands: []string{`ls ` + dstDir},
	})
	if err != nil {
		t.Error(err)
	}
	// After clone(), only the student repo is present in dstDir.
	// The tests and assignments repos are mounted directly into the container as read-only bind mounts.
	if !strings.Contains(out, qf.StudentRepoName(qfUserName)) {
		t.Errorf("expected %q to contain %q", out, qf.StudentRepoName(qfUserName))
	}
	for _, s := range []string{qf.TestsRepo, qf.AssignmentsRepo} {
		if strings.Contains(out, s) {
			t.Errorf("expected %q not to contain %q (should be a read-only mount, not a copy)", out, s)
		}
	}
}

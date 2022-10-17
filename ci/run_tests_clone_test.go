package ci

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/quickfeed/quickfeed/internal/rand"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

func init() {
	os.Setenv("QUICKFEED_REPOSITORY_PATH", "$HOME/tmp/courses")
}

func TestCloneAndCopyRunTests(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	sc, qfUserName := scm.GetTestUserAndSCM(t)

	dstDir := t.TempDir()

	course := &qf.Course{
		Code:             "QF101",
		Provider:         "github",
		OrganizationName: qfTestOrg,
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
		Organization: course.GetOrganizationName(),
		Repository:   qf.AssignmentsRepo,
		DestDir:      course.CloneDir(),
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(clonedAssignmentsRepo)

	clonedTestsRepo, err := sc.Clone(ctx, &scm.CloneOptions{
		Organization: course.GetOrganizationName(),
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
	for _, s := range []string{qf.TestsRepo, qf.AssignmentsRepo, qf.StudentRepoName(qfUserName)} {
		if !strings.Contains(out, s) {
			t.Errorf("expected %q to contain %q", out, s)
		}
	}
}

package scm_test

import (
	"context"
	"os"
	"testing"

	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
)

// This test will delete all open and closed issues for the test user and organization.
// The test is skipped unless run with: SCM_TESTS=1 go test -v -run TestDeleteAllIssues
func TestDeleteAllIssues(t *testing.T) {
	if os.Getenv("SCM_TESTS") == "" {
		t.SkipNow()
	}
	qfTestOrg := scm.GetTestOrganization(t)
	s, qfTestUser := scm.GetTestSCM(t)

	ctx := context.Background()
	opt := &scm.RepositoryOptions{
		Owner: qfTestOrg,
		Path:  qf.StudentRepoName(qfTestUser),
	}
	if err := s.DeleteIssues(ctx, opt); err != nil {
		t.Fatal(err)
	}
}

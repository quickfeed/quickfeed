package assignments

import (
	"context"
	"os"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/scm"
	"go.uber.org/zap"
)

// To run this test, please see instructions in the developer guide (dev.md).

func TestFetchAssignments(t *testing.T) {
	qfTestOrg := os.Getenv("QF_TEST_ORG")
	if len(qfTestOrg) < 1 {
		qfTestOrg = "qf101"
		t.Logf("This test requires access to the '%s' GitHub organization; to use another organization set the 'QF_TEST_ORG' environment variable", qfTestOrg)
	}
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if len(accessToken) < 1 {
		t.Skipf("This test requires that 'GITHUB_ACCESS_TOKEN' is set and that you have access to the '%v' GitHub organization", qfTestOrg)
	}

	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), "github", accessToken)
	if err != nil {
		t.Fatal(err)
	}

	course := &pb.Course{
		Name:             "QuickFeed Test Course",
		OrganizationPath: qfTestOrg,
	}

	assignments, err := FetchAssignments(context.Background(), s, course)
	if err != nil {
		t.Fatal(err)
	}
	// We don't actually test anything here since we don't know how many assignments are in QF_TEST_ORG
	for _, assignment := range assignments {
		t.Logf("assignment: %v", assignment)
	}
}

package assignments

import (
	"context"
	"testing"

	pb "github.com/autograde/quickfeed/ag"
	"github.com/autograde/quickfeed/scm"
	"go.uber.org/zap"
)

// To run this test, please see instructions in the developer guide (dev.md).

func TestFetchAssignments(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	accessToken := scm.GetAccessToken(t)

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

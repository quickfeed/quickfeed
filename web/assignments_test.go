package web

import (
	"context"
	"os"
	"testing"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
	"go.uber.org/zap"
)

const (
	gitHubTestOrg   = "autograder-test"
	gitHubTestOrgID = 30462712
)

// To enable this test, please see instructions in the developer guide (dev.md).
// You will also need access to the autograder-test organization; you may request
// access by sending your GitHub username to hein.meling at uis.no.

func TestFetchAssignments(t *testing.T) {
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if len(accessToken) < 1 {
		t.Skip("This test requires a 'GITHUB_ACCESS_TOKEN' and access to the 'autograder-test' GitHub organization")
	}
	provider := "github"

	var s scm.SCM
	s, err := scm.NewSCMClient(zap.NewNop().Sugar(), provider, accessToken)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	courseOrgID := uint64(gitHubTestOrgID)
	if courseOrgID == 0 {
		// find course directory ID for 'autograder-test' or your organization
		orgs, err := s.ListOrganizations(ctx)
		if err != nil {
			t.Fatal(err)
		}
		for _, org := range orgs {
			if org.Path == gitHubTestOrg {
				courseOrgID = org.ID
				t.Logf("To speed up test; update const to 'gitHubTestOrgID = %v'", org.ID)
			}
		}
	}

	course := &pb.Course{
		Name:           "Autograder Test Course",
		OrganizationID: courseOrgID,
	}

	assignments, err := fetchAssignments(ctx, s, course)
	if err != nil {
		t.Fatal(err)
	}
	for _, assignment := range assignments {
		t.Logf("assignment: %v", assignment)
	}
}

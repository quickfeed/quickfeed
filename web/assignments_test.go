package web

import (
	"context"
	"os"
	"testing"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
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
	s, err := scm.NewSCMClient(provider, accessToken)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	courseDirID := uint64(gitHubTestOrgID)
	if courseDirID == 0 {
		// find course directory ID for 'autograder-test' or your organization
		dirs, err := s.ListOrganizations(ctx)
		if err != nil {
			t.Fatal(err)
		}
		for _, dir := range dirs {
			if dir.Path == gitHubTestOrg {
				courseDirID = dir.ID
				t.Logf("To speed up test; update const to 'gitHubTestOrgID = %v'", dir.ID)
			}
		}
	}

	course := &pb.Course{
		Name:           "Autograder Test Course",
		OrganizationID: courseDirID,
	}

	assignments, err := fetchAssignments(ctx, s, course)
	if err != nil {
		t.Fatal(err)
	}
	for _, assignment := range assignments {
		t.Logf("assignment: %v", assignment)
	}
}

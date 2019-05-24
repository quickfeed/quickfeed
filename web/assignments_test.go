package web_test

import (
	"context"
	"os"
	"testing"

	pb "github.com/autograde/aguis/ag"
	"github.com/autograde/aguis/scm"
	"github.com/autograde/aguis/web"
)

const (
	gitHubTestOrg   = "autograder-test"
	gitHubTestOrgID = 30462712
)

func TestFetchAssignments(t *testing.T) {
	accessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if len(accessToken) < 1 {
		t.Skip("This test requires a 'GITHUB_ACCESS_TOKEN' for the 'autograder-test' GitHub organization")
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
		dirs, err := s.ListDirectories(ctx)
		if err != nil {
			t.Fatal(err)
		}
		for _, dir := range dirs {
			if dir.Path == gitHubTestOrg {
				courseDirID = dir.Id
				t.Logf("To speed up test; update const to 'gitHubTestOrgID = %v'", dir.Id)
			}
		}
	}
	course := &pb.Course{
		Name:        "Autograder Test Course",
		DirectoryId: courseDirID,
	}

	assignments, err := web.FetchAssignments(ctx, s, course)
	if err != nil {
		t.Fatal(err)
	}
	for _, assignment := range assignments {
		t.Logf("assignment: %v", assignment)
	}
}

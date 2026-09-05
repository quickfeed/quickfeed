package scm_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"google.golang.org/protobuf/testing/protocmp"
)

const (
	qf101Org   = "qf101"
	qf101OrdID = 77283363
)

// To run this test, please see instructions in the developer guide (dev.md).

func TestGetOrganization(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	s, qfTestUser := scm.GetTestSCM(t)
	org, err := s.GetOrganization(context.Background(), &scm.OrganizationOptions{
		Name:     qfTestOrg,
		Username: qfTestUser,
	})
	if err != nil {
		t.Fatal(err)
	}
	if qfTestOrg == qf101Org {
		if org.GetScmOrganizationID() != qf101OrdID {
			t.Errorf("GetOrganization(%q) = %d, expected %d", qfTestOrg, org.GetScmOrganizationID(), qf101OrdID)
		}
	} else {
		// Otherwise, we just print the organization result
		t.Logf("org: %v", org)
	}
}

// Test case for Creating new Issue on a git Repository
// This test assumes the repositories are compared against assignments, and that
// user/group forks may be ahead by zero or more commits.
func TestCommitsAhead(t *testing.T) {
	qfTestOrg := scm.GetTestOrganization(t)
	s, _ := scm.GetTestSCM(t)

	tests := []struct {
		name      string
		opt       *scm.RepositoryOptions
		wantAhead int
		wantErr   bool
	}{
		{name: "CourseRepo", opt: &scm.RepositoryOptions{Repo: "tests", Owner: qfTestOrg}, wantErr: true},
		{name: "CourseRepoByID", opt: &scm.RepositoryOptions{ID: 328688692}, wantErr: true},
		{name: "CourseRepoInfo", opt: &scm.RepositoryOptions{Repo: "info", Owner: qfTestOrg}, wantErr: true},
		{name: "CourseRepoInfoByID", opt: &scm.RepositoryOptions{ID: 328688666}, wantErr: true},
		{name: "NonExistentRepo", opt: &scm.RepositoryOptions{Repo: "some-other-repo", Owner: qfTestOrg}, wantErr: true},
	}
	for _, tt := range tests {
		name := qtest.Name(tt.name, []string{"ID", "Owner", "Repo"}, tt.opt.ID, tt.opt.Owner, tt.opt.Repo)
		t.Run(name, func(t *testing.T) {
			ahead, err := s.CommitsAhead(context.Background(), tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommitsAhead(%+v) error = %v, wantErr %v", *tt.opt, err, tt.wantErr)
			}
			if ahead != tt.wantAhead {
				t.Errorf("CommitsAhead(%+v) = %d, want %d", *tt.opt, ahead, tt.wantAhead)
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	s := scm.NewMockedGithubSCMClient(qtest.Logger(t),
		scm.WithMockOrgs("meling"), // user "meling" with ID 1
		scm.WithMembers(github.Membership{
			User: &github.User{
				ID:        github.Int64(2),
				Login:     new("avatar_user"),
				AvatarURL: new("https://avatar.com"),
			},
		}),
	)
	ctx := t.Context()

	// Test successfully fetching a user
	gotUser, err := s.GetUserByID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	wantUser := &qf.User{
		Login:       "meling",
		ScmRemoteID: 1,
	}
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("GetUserByID() mismatch (-want +got):\n%s", diff)
	}

	// Test successfully fetching a user with AvatarURL
	gotUser, err = s.GetUserByID(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	wantUser = &qf.User{
		Login:       "avatar_user",
		AvatarURL:   "https://avatar.com",
		ScmRemoteID: 2,
	}
	if diff := cmp.Diff(wantUser, gotUser, protocmp.Transform()); diff != "" {
		t.Errorf("GetUserByID() mismatch (-want +got):\n%s", diff)
	}

	// Test handling errors when the user doesn't exist
	_, err = s.GetUserByID(ctx, 999)
	if err == nil {
		t.Error("expected error for non-existent user ID 999")
	}
}

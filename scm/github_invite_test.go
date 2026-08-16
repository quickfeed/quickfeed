package scm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-github/v62/github"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
)

func restoreInviteWait(t *testing.T) {
	t.Helper()
	attempts, initial, maxDelay := waitForInviteMaxAttempts, waitForInviteInitialDelay, waitForInviteMaxDelay
	t.Cleanup(func() {
		waitForInviteMaxAttempts = attempts
		waitForInviteInitialDelay = initial
		waitForInviteMaxDelay = maxDelay
	})
}

func TestAcceptOrgInvitationRetriesAccepted(t *testing.T) {
	restoreInviteWait(t)
	waitForInviteMaxAttempts = 5
	waitForInviteInitialDelay = time.Millisecond
	waitForInviteMaxDelay = time.Millisecond

	members := []github.Membership{
		{Organization: &ghOrgFoo, User: &meling, State: new("pending")},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo), WithMembers(members...), WithInvitationNotReady(2))

	if err := s.acceptOrgInvitation(context.Background(), &InvitationOptions{
		Login:       "meling",
		Owner:       "foo",
		AccessToken: "dummy",
	}); err != nil {
		t.Fatalf("acceptOrgInvitation() after retries: %v", err)
	}
}

func TestAcceptOrgInvitationGivesUp(t *testing.T) {
	restoreInviteWait(t)
	waitForInviteMaxAttempts = 3
	waitForInviteInitialDelay = time.Millisecond
	waitForInviteMaxDelay = time.Millisecond

	members := []github.Membership{
		{Organization: &ghOrgFoo, User: &meling, State: new("pending")},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo), WithMembers(members...), WithInvitationNotReady(10))

	err := s.acceptOrgInvitation(context.Background(), &InvitationOptions{
		Login:       "meling",
		Owner:       "foo",
		AccessToken: "dummy",
	})
	if err == nil {
		t.Fatal("acceptOrgInvitation() succeeded, want error after exhausting retries")
	}
	if !strings.Contains(err.Error(), "invitation not ready after 3 attempts") {
		t.Errorf("acceptOrgInvitation() error = %v, want exhausted-retries message", err)
	}
	if !isGitHubAccepted(err) {
		t.Errorf("acceptOrgInvitation() error should unwrap to AcceptedError, got %v", err)
	}
}

func TestAcceptOrgInvitationDoesNotRetryOtherErrors(t *testing.T) {
	restoreInviteWait(t)
	waitForInviteMaxAttempts = 5
	waitForInviteInitialDelay = 50 * time.Millisecond
	waitForInviteMaxDelay = 50 * time.Millisecond

	// Org exists but has no membership, so PATCH returns 404, not 202.
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgFoo))

	start := time.Now()
	err := s.acceptOrgInvitation(context.Background(), &InvitationOptions{
		Login:       "meling",
		Owner:       "foo",
		AccessToken: "dummy",
	})
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("acceptOrgInvitation() succeeded, want not-found error")
	}
	if isGitHubAccepted(err) {
		t.Errorf("acceptOrgInvitation() wrapped AcceptedError, want a hard error: %v", err)
	}
	if elapsed >= 50*time.Millisecond {
		t.Errorf("acceptOrgInvitation() retried a non-accepted error; took %v", elapsed)
	}
}

func TestAcceptOrgInvitationInvalidOptions(t *testing.T) {
	s := NewMockedGithubSCMClient(qtest.Logger(t))
	if err := s.acceptOrgInvitation(context.Background(), &InvitationOptions{}); err == nil {
		t.Fatal("acceptOrgInvitation() succeeded, want invalid options error")
	}
}

func TestUpdateEnrollmentRetriesInvitation(t *testing.T) {
	restoreInviteWait(t)
	waitForInviteMaxAttempts = 5
	waitForInviteInitialDelay = time.Millisecond
	waitForInviteMaxDelay = time.Millisecond

	g := map[string]map[string][]github.User{
		"bar": {
			"assignments": {},
			"frank-labs":  {},
		},
	}
	s := NewMockedGithubSCMClient(qtest.Logger(t), WithOrgs(ghOrgBar), WithRepos(repos...), WithGroups(g), WithInvitationNotReady(2))
	got, err := s.UpdateEnrollment(context.Background(), &UpdateEnrollmentOptions{
		Organization: "bar",
		User:         "frank",
		Status:       qf.Enrollment_STUDENT,
		AccessToken:  "dummy",
	})
	if err != nil {
		t.Fatalf("UpdateEnrollment() after invitation retries: %v", err)
	}
	want := &Repository{Owner: "bar", Repo: "frank-labs"}
	if diff := cmp.Diff(want, got, cmpopts.IgnoreFields(Repository{}, "ID")); diff != "" {
		t.Errorf("UpdateEnrollment() mismatch (-want +got):\n%s", diff)
	}
}

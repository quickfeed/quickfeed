package interceptor_test

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
)

func TestRefreshTokens(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin", "user"),
		web.WithInterceptors(
			web.UserInterceptorFunc,
			web.TokenInterceptorFunc,
		),
	)
	tm := client.TokenManager()

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "Admin User", Login: "admin", ScmRemoteID: 1})
	user := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "Test User", Login: "user", ScmRemoteID: 2})

	course := &qf.Course{
		ID:                  1,
		ScmOrganizationID:   1,
		ScmOrganizationName: qtest.MockOrg,
	}
	group := &qf.Group{
		ID:       1,
		Name:     "test",
		CourseID: 1,
		Users:    []*qf.User{user},
	}
	qtest.CreateCourse(t, db, admin, course)
	qtest.EnrollStudent(t, db, user, course)

	adminCtx := client.Context(t, admin)
	userCookie := client.Cookie(t, user)

	operations := []struct {
		name      string
		operation func() error
		wantUser  bool // whether user should be in token update list after operation
	}{
		{
			name:      "initial state should not trigger updates",
			operation: func() error { return nil },
			wantUser:  false,
		},
		{
			name: "read operations don't trigger updates",
			operation: func() error {
				_, err := client.GetUsers(adminCtx, &qf.Void{})
				return err
			},
			wantUser: false,
		},
		{
			name: "user update triggers token refresh",
			operation: func() error {
				_, err := client.UpdateUser(adminCtx, user)
				return err
			},
			wantUser: true,
		},
		{
			name: "create group doesn't trigger update",
			operation: func() error {
				_, err := client.CreateGroup(adminCtx, group)
				return err
			},
			wantUser: false,
		},
		{
			name: "update group triggers token refresh",
			operation: func() error {
				_, err := client.UpdateGroup(adminCtx, group)
				return err
			},
			wantUser: true,
		},
		{
			name: "delete group triggers token refresh",
			operation: func() error {
				_, err := client.DeleteGroup(adminCtx, &qf.GroupRequest{
					GroupID:  group.GetID(),
					CourseID: course.GetID(),
				})
				return err
			},
			wantUser: true,
		},
	}

	for _, op := range operations {
		t.Run(op.name, func(t *testing.T) {
			if err := op.operation(); err != nil {
				t.Error(err)
			}
			checkUpdateTokenRequired(t, tm, user.GetID(), op.wantUser, "user")
			checkUpdateTokenRequired(t, tm, admin.GetID(), false, "admin") // admin should never need token update in this test

			updatedCookie := getUserRefreshCookie(t, client, userCookie)
			if op.wantUser {
				if updatedCookie == "" {
					t.Fatal("expected refreshed cookie in response header")
				}
				userCookie = updatedCookie
				checkUpdateTokenRequired(t, tm, user.GetID(), false, "user")
			} else if updatedCookie != "" {
				t.Fatal("unexpected refreshed cookie in response header")
			}
		})
	}
}

func getUserRefreshCookie(t *testing.T, client *web.MockClient, cookie string) string {
	t.Helper()

	userCtx, userInfo := connect.NewClientContext(t.Context())
	userInfo.RequestHeader().Set(auth.Cookie, cookie)
	if _, err := client.GetUser(userCtx, &qf.Void{}); err != nil {
		t.Fatal(err)
	}
	if got := userInfo.RequestHeader().Get(auth.Cookie); got != cookie {
		t.Fatal("request context cookie should not be auto-updated")
	}
	return userInfo.ResponseHeader().Get(auth.SetCookie)
}

func checkUpdateTokenRequired(t *testing.T, tm *auth.TokenManager, userID uint64, expected bool, userType string) {
	t.Helper()

	user, err := tm.Database().GetUser(userID)
	if err != nil {
		t.Fatal(err)
	}
	if user.GetUpdateToken() != expected {
		if expected {
			t.Errorf("%s token should be updated but is not", userType)
		} else {
			t.Errorf("%s token should not be updated but is", userType)
		}
	}
}

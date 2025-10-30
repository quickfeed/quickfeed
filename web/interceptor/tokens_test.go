package interceptor_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
)

func TestRefreshTokens(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()

	client := web.NewMockClient(t, db, scm.WithMockOrgs("admin"),
		web.WithInterceptors(
			web.UserInterceptorFunc,
			web.TokenInterceptorFunc,
		),
	)
	tm := client.TokenManager()
	ctx := t.Context()

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "admin", Login: "admin"})
	user := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "user", Login: "user"})
	adminCookie, adminClaims := createUserAuth(t, tm, admin.GetID(), true)
	userCookie, userClaims := createUserAuth(t, tm, user.GetID(), false)

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
				_, err := client.GetUsers(ctx, qtest.RequestWithCookie(&qf.Void{}, adminCookie))
				return err
			},
			wantUser: false,
		},
		{
			name: "user update triggers token refresh",
			operation: func() error {
				_, err := client.UpdateUser(ctx, qtest.RequestWithCookie(user, adminCookie))
				return err
			},
			wantUser: true,
		},
		{
			name: "create group doesn't trigger update",
			operation: func() error {
				_, err := client.CreateGroup(ctx, qtest.RequestWithCookie(group, adminCookie))
				return err
			},
			wantUser: false,
		},
		{
			name: "update group triggers token refresh",
			operation: func() error {
				_, err := client.UpdateGroup(ctx, qtest.RequestWithCookie(group, adminCookie))
				return err
			},
			wantUser: true,
		},
		{
			name: "delete group triggers token refresh",
			operation: func() error {
				_, err := client.DeleteGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{
					GroupID:  group.GetID(),
					CourseID: course.GetID(),
				}, adminCookie))
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
			checkTokenUpdateRequired(t, tm, userClaims, op.wantUser, "user")
			checkTokenUpdateRequired(t, tm, adminClaims, false, "admin") // admin should never need token update in this test

			if op.wantUser {
				// If user token needs update, simulate token refresh
				if _, err := client.GetUser(ctx, qtest.RequestWithCookie(&qf.Void{}, userCookie)); err != nil {
					t.Error(err)
				}
				checkTokenUpdateRequired(t, tm, userClaims, false, "user")
			}
		})
	}
}

// createUserAuth returns an authentication cookie and JWT claims for a user.
func createUserAuth(t *testing.T, tm *auth.TokenManager, userID uint64, isAdmin bool) (string, *auth.Claims) {
	t.Helper()
	cookie, err := tm.NewAuthCookie(userID)
	if err != nil {
		t.Fatal(err)
	}
	claims := &auth.Claims{
		UserID: userID,
		Admin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Minute)),
		},
	}
	return cookie.String(), claims
}

func checkTokenUpdateRequired(t *testing.T, tm *auth.TokenManager, claims *auth.Claims, expected bool, userType string) {
	t.Helper()
	updated, err := tm.UpdateCookie(claims)
	if err != nil {
		t.Error(err)
	}
	if (updated != nil) != expected {
		if expected {
			t.Errorf("%s token should be updated but is not", userType)
		} else {
			t.Errorf("%s token should not be updated but is", userType)
		}
	}
}

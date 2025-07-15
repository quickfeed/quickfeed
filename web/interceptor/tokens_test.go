package interceptor_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
)

func TestRefreshTokens(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qtest.Logger(t)

	tm, err := auth.NewTokenManager(db)
	if err != nil {
		t.Fatal(err)
	}
	client := web.MockClient(t, db, scm.WithMockOrgs("admin"), connect.WithInterceptors(
		interceptor.NewUserInterceptor(logger, tm),
		interceptor.NewTokenInterceptor(tm),
	))
	ctx := context.Background()

	f := func(t *testing.T, id uint64) string {
		cookie, err := tm.NewAuthCookie(id)
		if err != nil {
			t.Fatal(err)
		}
		return cookie.String()
	}

	admin := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "admin", Login: "admin"})
	user := qtest.CreateFakeCustomUser(t, db, &qf.User{Name: "user", Login: "user"})
	adminCookie := f(t, admin.GetID())
	userCookie := f(t, user.GetID())
	adminClaims := &auth.Claims{
		UserID: admin.GetID(),
		Admin:  true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Minute)),
		},
	}
	userClaims := &auth.Claims{
		UserID: user.GetID(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Minute)),
		},
	}
	if updateRequired(t, tm, userClaims) || updateRequired(t, tm, adminClaims) {
		t.Error("No users should be in the token update list at the start")
	}
	if _, err := client.GetUsers(ctx, qtest.RequestWithCookie(&qf.Void{}, adminCookie)); err != nil {
		t.Fatal(err)
	}
	if updateRequired(t, tm, adminClaims) || updateRequired(t, tm, userClaims) {
		t.Error("No users should be in the token update list")
	}
	if _, err := client.UpdateUser(ctx, qtest.RequestWithCookie(user, adminCookie)); err != nil {
		t.Fatal(err)
	}
	if !updateRequired(t, tm, userClaims) {
		t.Error("User must be in the token update list after admin has updated the user's information")
	}
	if _, err := client.GetUser(ctx, qtest.RequestWithCookie(&qf.Void{}, userCookie)); err != nil {
		t.Fatal(err)
	}
	if updateRequired(t, tm, userClaims) {
		t.Error("User should not be in the token update list after the token has been updated")
	}
	course := &qf.Course{
		ID:                  1,
		ScmOrganizationID:   1,
		ScmOrganizationName: qtest.MockOrg,
	}
	group := &qf.Group{
		ID:       1,
		Name:     "test",
		CourseID: 1,
		Users: []*qf.User{
			user,
		},
	}
	qtest.CreateCourse(t, db, admin, course)
	qtest.EnrollStudent(t, db, user, course)
	if _, err := client.CreateGroup(ctx, qtest.RequestWithCookie(group, adminCookie)); err != nil {
		t.Fatal(err)
	}
	if updateRequired(t, tm, userClaims) {
		t.Error("User should not be in the token update list after methods that don't affect the user's information")
	}
	if _, err := client.UpdateGroup(ctx, qtest.RequestWithCookie(group, adminCookie)); err != nil {
		t.Fatal(err)
	}
	if !updateRequired(t, tm, userClaims) {
		t.Error("User must be in the token update group after changes to the group")
	}
	if _, err := client.GetUser(ctx, qtest.RequestWithCookie(&qf.Void{}, userCookie)); err != nil {
		t.Fatal(err)
	}
	if updateRequired(t, tm, userClaims) {
		t.Error("User should be removed from the token update list after the user's token has been updated")
	}
	if _, err := client.DeleteGroup(ctx, qtest.RequestWithCookie(&qf.GroupRequest{
		GroupID:  group.GetID(),
		CourseID: course.GetID(),
	}, adminCookie)); err != nil {
		t.Fatal(err)
	}
	if !updateRequired(t, tm, userClaims) {
		t.Error("User must be in the token update list after the group has been deleted")
	}
	if updateRequired(t, tm, adminClaims) {
		t.Error("Admin should not be in the token update list")
	}
}

func updateRequired(t *testing.T, tm *auth.TokenManager, claims *auth.Claims) bool {
	t.Helper()
	updated, err := tm.UpdateCookie(claims)
	if err != nil {
		t.Error(err)
	}
	return updated != nil
}

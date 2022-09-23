package interceptor_test

import (
	"context"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/golang-jwt/jwt"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
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
	serveFn, shutdown := web.MockQuickFeedServer(t, logger, db, connect.WithInterceptors(
		interceptor.UnaryUserVerifier(logger, tm),
		interceptor.TokenRefresher(tm),
	))
	go serveFn()

	client := qtest.QuickFeedClient("")

	ctx := context.Background()
	f := func(t *testing.T, id uint64) string {
		cookie, err := tm.NewAuthCookie(id)
		if err != nil {
			t.Fatal(err)
		}
		return cookie.String()
	}

	admin := qtest.CreateFakeUser(t, db, 1)
	user := qtest.CreateFakeUser(t, db, 56)
	adminCookie := f(t, admin.ID)
	userCookie := f(t, user.ID)
	adminClaims := &auth.Claims{
		UserID: admin.ID,
		Admin:  true,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(1 * time.Minute).Unix(),
		},
	}
	userClaims := &auth.Claims{
		UserID: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(1 * time.Minute).Unix(),
		},
	}
	if tm.UpdateRequired(adminClaims) || tm.UpdateRequired(userClaims) {
		t.Error("No users should be in the token update list at the start")
	}
	if _, err := client.GetUsers(ctx, requestWithCookie(&qf.Void{}, adminCookie)); err != nil {
		t.Fatal(err)
	}
	if tm.UpdateRequired(adminClaims) || tm.UpdateRequired(userClaims) {
		t.Error("No users should be in the token update list")
	}
	if _, err := client.UpdateUser(ctx, requestWithCookie(user, adminCookie)); err != nil {
		t.Fatal(err)
	}
	if !tm.UpdateRequired(userClaims) {
		t.Error("User must be in the token update list after admin has updated the user's information")
	}
	if _, err := client.GetUser(ctx, requestWithCookie(&qf.Void{}, userCookie)); err != nil {
		t.Fatal(err)
	}
	if tm.UpdateRequired(userClaims) {
		t.Error("User should not be in the token update list after the token has been updated")
	}
	course := &qf.Course{
		ID:               1,
		OrganizationID:   1,
		OrganizationPath: "test",
		Provider:         "fake",
	}
	group := &qf.Group{
		ID:       1,
		Name:     "test",
		CourseID: 1,
		Users: []*qf.User{
			user,
		},
	}
	if _, err := client.CreateCourse(ctx, requestWithCookie(course, adminCookie)); err != nil {
		t.Fatal(err)
	}
	if !tm.UpdateRequired(adminClaims) {
		t.Error("Admin must be in the token update list after creating a new course")
	}
	qtest.EnrollStudent(t, db, user, course)
	if _, err := client.CreateGroup(ctx, requestWithCookie(group, adminCookie)); err != nil {
		t.Fatal(err)
	}
	if tm.UpdateRequired(userClaims) {
		t.Error("User should not be in the token update list after methods that don't affect the user's information")
	}
	if _, err := client.UpdateGroup(ctx, requestWithCookie(group, adminCookie)); err != nil {
		t.Fatal(err)
	}
	if !tm.UpdateRequired(userClaims) {
		t.Error("User must be in the token update group after changes to the group")
	}
	if _, err := client.GetUser(ctx, requestWithCookie(&qf.Void{}, userCookie)); err != nil {
		t.Fatal(err)
	}
	if tm.UpdateRequired(userClaims) {
		t.Error("User should be removed from the token update list after the user's token has been updated")
	}
	if _, err := client.DeleteGroup(ctx, requestWithCookie(&qf.GroupRequest{
		GroupID:  group.ID,
		CourseID: course.ID,
	}, adminCookie)); err != nil {
		t.Fatal(err)
	}
	if !tm.UpdateRequired(userClaims) {
		t.Error("User must be in the token update list after the group has been deleted")
	}
	if tm.UpdateRequired(adminClaims) {
		t.Error("Admin should not be in the token update list")
	}
	shutdown(ctx)
}

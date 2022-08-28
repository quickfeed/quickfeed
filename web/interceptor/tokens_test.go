package interceptor_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/golang-jwt/jwt"
	"github.com/quickfeed/quickfeed/ci"
	"github.com/quickfeed/quickfeed/internal/qtest"
	"github.com/quickfeed/quickfeed/qf"
	"github.com/quickfeed/quickfeed/qf/qfconnect"
	"github.com/quickfeed/quickfeed/qlog"
	"github.com/quickfeed/quickfeed/scm"
	"github.com/quickfeed/quickfeed/web"
	"github.com/quickfeed/quickfeed/web/auth"
	"github.com/quickfeed/quickfeed/web/interceptor"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func TestRefreshTokens(t *testing.T) {
	db, cleanup := qtest.TestDB(t)
	defer cleanup()
	logger := qlog.Logger(t)
	ags := web.NewQuickFeedService(logger.Desugar(), db, scm.TestSCMManager(), web.BaseHookOptions{}, &ci.Local{})

	tm, err := auth.NewTokenManager(db, "test")
	if err != nil {
		t.Fatal(err)
	}

	interceptors := connect.WithInterceptors(
		interceptor.UnaryUserVerifier(logger, tm),
		interceptor.TokenRefresher(tm),
	)

	router := http.NewServeMux()
	router.Handle(qfconnect.NewQuickFeedServiceHandler(ags, interceptors))
	muxServer := &http.Server{
		Handler:           h2c.NewHandler(router, &http2.Server{}),
		Addr:              "127.0.0.1:8081",
		ReadHeaderTimeout: 3 * time.Second, // to prevent Slowloris (CWE-400)
	}
	go func() {
		if err := muxServer.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				t.Errorf("Server exited with unexpected error: %v", err)
			}
			return
		}
	}()

	ctx := context.Background()
	client := qtest.QuickFeedClient("")
	f := func(t *testing.T, id uint64) context.Context {
		token, err := tm.NewAuthCookie(id)
		if err != nil {
			t.Fatal(err)
		}
		return qtest.WithAuthCookie(ctx, token)
	}

	admin := qtest.CreateFakeUser(t, db, 1)
	user := qtest.CreateFakeUser(t, db, 56)
	adminCtx := f(t, admin.ID)
	userCtx := f(t, user.ID)
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
	if _, err := client.GetUsers(adminCtx, connect.NewRequest(&qf.Void{})); err != nil {
		t.Fatal(err)
	}
	if tm.UpdateRequired(adminClaims) || tm.UpdateRequired(userClaims) {
		t.Error("No users should be in the token update list")
	}
	if _, err := client.UpdateUser(adminCtx, connect.NewRequest(user)); err != nil {
		t.Fatal(err)
	}
	if !tm.UpdateRequired(userClaims) {
		t.Error("User must be in the token update list after admin has updated the user's information")
	}
	if _, err := client.GetUser(userCtx, connect.NewRequest(&qf.Void{})); err != nil {
		t.Fatal(err)
	}
	if tm.UpdateRequired(userClaims) {
		t.Error("User should not be in the token update list after the token has been updated")
	}
	course := &qf.Course{
		ID:               1,
		OrganizationID:   1,
		OrganizationPath: "testorg",
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
	if _, err := client.CreateCourse(adminCtx, connect.NewRequest(course)); err != nil {
		t.Fatal(err)
	}
	if !tm.UpdateRequired(adminClaims) {
		t.Error("Admin must be in the token update list after creating a new course")
	}
	qtest.EnrollStudent(t, db, user, course)
	if _, err := client.CreateGroup(adminCtx, connect.NewRequest(group)); err != nil {
		t.Fatal(err)
	}
	if tm.UpdateRequired(userClaims) {
		t.Error("User should not be in the token update list after methods that don't affect the user's information")
	}
	if _, err := client.UpdateGroup(adminCtx, connect.NewRequest(group)); err != nil {
		t.Fatal(err)
	}
	if !tm.UpdateRequired(userClaims) {
		t.Error("User must be in the token update group after changes to the group")
	}
	if _, err := client.GetUser(userCtx, connect.NewRequest(&qf.Void{})); err != nil {
		t.Fatal(err)
	}
	if tm.UpdateRequired(userClaims) {
		t.Error("User should be removed from the token update list after the user's token has been updated")
	}
	if _, err := client.DeleteGroup(adminCtx, connect.NewRequest(&qf.GroupRequest{
		GroupID:  group.ID,
		CourseID: course.ID,
	})); err != nil {
		t.Fatal(err)
	}
	if !tm.UpdateRequired(userClaims) {
		t.Error("User must be in the token update list after the group has been deleted")
	}
	if tm.UpdateRequired(adminClaims) {
		t.Error("Admin should not be in the token update list")
	}
	if err = muxServer.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
}
